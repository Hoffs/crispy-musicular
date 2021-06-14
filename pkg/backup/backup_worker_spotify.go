package backup

import (
	"context"
	"math"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/syncplus"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func (b *backuper) backupSpotify(ctx context.Context, state *backupState, authState *auth.State) (err error) {
	// There should be no long term issues with this as refresh token doesn't change on subsequent
	// authorizations and probably only changes if auth is revoked for the app or something is reset
	// by Spotify.
	sAuth := spotify.NewAuthenticator("", spotify.ScopePlaylistReadPrivate)
	state.spotify = sAuth.NewClient(&oauth2.Token{RefreshToken: authState.RefreshToken})
	state.spotify.AutoRetry = true // Auto retry on rate limit

	usr, err := state.spotify.CurrentUser()
	if err != nil {
		log.Error().Err(err).Msg("backuper: failed to get current user, is refresh token invalid?")
		return
	}

	state.bp, err = b.createBackup(usr.ID)
	if err != nil {
		log.Error().Err(err).Msg("backuper: could not create backup entry")
		return
	}
	workers := b.config.WorkerCount
	log.Info().Msgf("backuper: starting backup for %s with %d workers", usr.ID, workers)

	// Use either 51 or if worker amount is higher, worker count + 1,
	// so that in best case scenario we prefetch enough data to saturate all workers.
	// Would maybe make sense to add also an upper bound with math.Min, so that the buffer size would
	// be too big.
	bufferSize := int(math.Max(float64(51), float64(workers+1)))
	pch := make(chan *spotify.SimplePlaylist, bufferSize)

	errs := make(map[uint8]error)

	for i := uint8(0); i < workers; i++ {
		state.wg.Add(1)
		go func(id uint8) {
			err := b.worker(state, pch)
			log.Debug().Err(err).Msgf("worker %d ended", id)
			if err != nil {
				errs[id] = err
			}
		}(i)
	}

	limit := 50 // Max playlists per page
	playlists, err := state.spotify.CurrentUsersPlaylistsOpt(&spotify.Options{Limit: &limit})
	if err != nil {
		log.Error().Err(err).Msg("backuper: couldn't get initial user playlists")
		return
	}

	for {
		log.Debug().Msgf("backuper: got playlist page, offset %d, limit %d, total %d", playlists.Offset, playlists.Limit, playlists.Total)

		for id := range playlists.Playlists {
			// use from array, since value from range function changes, but not the pointer
			p := &playlists.Playlists[id]

			if !b.shouldSavePlaylist(usr.ID, p) {
				log.Debug().Msgf("backuper: skipping '%s' with id '%s'", p.Name, p.ID)
				continue
			}

			log.Debug().Msgf("backuper: sending '%s' to worker with pointer %p", p.Name, p)

			// New struct is created, so sending pointer causes no issues even if next page is loaded before
			// previous page has finished.
			pch <- p
		}

		err = state.spotify.NextPage(playlists)
		if err == spotify.ErrNoMorePages {
			err = nil
			break
		}

		if err != nil {
			// This might leave some stuff running
			close(pch)
			return
		}
	}

	// Close to trigger end of work queue
	close(pch)

	timedOut := syncplus.WaitContext(ctx, &state.wg)
	if timedOut {
		log.Warn().Msg("backuper: workers did not finish in time")
	}

	for id, idErr := range errs {
		if idErr != nil {
			err = idErr
			log.Error().Err(idErr).Msgf("backuper: worker %d errored", id)
		}
	}

	return
}

// listens for playlists on channel
func (b *backuper) worker(st *backupState, playlists <-chan *spotify.SimplePlaylist) (err error) {
	defer st.wg.Done()

	for {
		select {
		case p := <-playlists:
			if p == nil {
				log.Debug().Msgf("backuper_worker: received nil playlist, exiting")
				return
			}

			log.Debug().Msgf("backuper_worker: received playlist '%s' with pointer '%p'", p.Name, p)
			if err != nil {
				log.Warn().Err(err).Msgf("backuper_worker: skipping playlist '%s' because error'ed already", p.Name)
				continue
			}

			err = b.savePlaylist(st, p)
			if err != nil {
				// don't exit, try to save other playlists
				log.Error().Err(err).Msgf("backuper_worker: encountered an error while saving playlist '%s'", p.Name)
			}
		case <-st.ctx.Done():
			log.Debug().Msg("backuper_worker: exiting")
			return
		}
	}
}

func (b *backuper) savePlaylist(st *backupState, playlist *spotify.SimplePlaylist) (err error) {
	// This already has default limit as max, so no need for options
	tracks, err := st.spotify.GetPlaylistTracks(playlist.ID)
	if err != nil {
		log.Error().Err(err).Msgf("backuper_worker: failed to get initial playlist tracks for '%s'", playlist.Name)
		return
	}

	p, err := b.addSpotifyPlaylist(st.bp, playlist)
	if err != nil {
		log.Error().Err(err).Msgf("backuper_worker: could not create playlist entry for '%s'", playlist.Name)
		return
	}

	for {
		log.Debug().Msgf("backuper_worker: got track page for '%s', offset %d, limit %d, total %d", playlist.Name, tracks.Offset, tracks.Limit, tracks.Total)

		for _, t := range tracks.Tracks {
			log.Debug().Msgf("backuper_worker: playlist '%s', track '%s'", playlist.Name, t.Track.Name)

			err = b.addSpotifyTrack(st.bp, p, &t)
			if err != nil {
				log.Error().Err(err).Msgf("backuper_worker: could not create track entry for '%s'/'%s'", playlist.Name, t.Track.Name)
				return
			}
		}

		err = st.spotify.NextPage(tracks)
		if err == spotify.ErrNoMorePages {
			return nil
		}

		if err != nil {
			return
		}
	}
}

// applies all rules in order, by default => should save
// 1. checks if IgnoreNotOwnedPlaylists and OwnerID != UserId, if true => shouldnt
// 2. checks if IgnoreOwnedPlaylists and OwnerID == UserId, if true => shouldnt
// 3. checks if exists in IgnoredPlaylistIds, if exists => shouldn't
// 4. checks if exists in SavedPlaylistIds, if exists => should
// this allows to have following posibilities,
// A. can ignore all not user created playlists
// B. can ignore all user created playlists
// C. can force save some not user created playlists
// D. can ignore some user (or not user if first option is false) created playlists
// E. can ignore all playlists (1 + 2) and only backup select few (3)
func (b *backuper) shouldSavePlaylist(userId string, p *spotify.SimplePlaylist) (save bool) {
	save = true
	if b.config.IgnoreNotOwnedPlaylists && p.Owner.ID != userId {
		save = false
	}

	if b.config.IgnoreOwnedPlaylists && p.Owner.ID == userId {
		save = false
	}

	id := string(p.ID)

	for _, savedId := range b.config.IgnoredPlaylistIds {
		save = save && (savedId != id)
	}

	for _, savedId := range b.config.SavedPlaylistIds {
		save = save || (savedId == id)
	}

	return
}
