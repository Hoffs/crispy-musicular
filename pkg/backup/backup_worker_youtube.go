package backup

import (
	"context"
	"math"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/syncplus"
	"github.com/hoffs/crispy-musicular/pkg/youtube"
	"github.com/rs/zerolog/log"
	gyoutube "google.golang.org/api/youtube/v3"
)

func (b *backuper) backupYoutube(ctx context.Context, state *backupState, authState *auth.State) (err error) {
	if authState.YoutubeRefreshToken == "" {
		log.Info().Msg("backuper: youtube account is not configured")
		return nil
	}

	auth := youtube.NewAuthenticator(b.config.YoutubeId, b.config.YoutubeSecret, b.config.YoutubeCallback)
	state.youtube, err = auth.FromRefreshToken(authState.YoutubeRefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("backuper: failed to create youtube service")
		return
	}

	// instead of .Do(), theres pretty cool Pages() which can call a function for every page.
	playlists, err := state.youtube.Playlists.List([]string{"snippet"}).Id(b.config.YoutubeSavedPlaylistIds...).MaxResults(50).Do()
	if err != nil {
		log.Error().Err(err).Msg("backuper: failed to get youtube playlists")
		return
	}

	workers := b.config.WorkerCount
	log.Info().Msgf("backuper: starting youtube backup with %d workers", workers)

	bufferSize := int(math.Max(float64(51), float64(workers+1)))
	pch := make(chan *gyoutube.Playlist, bufferSize)

	errs := make(map[uint8]error)

	for i := uint8(0); i < workers; i++ {
		state.wg.Add(1)
		go func(id uint8) {
			err := b.worker_youtube(state, pch)
			log.Debug().Err(err).Msgf("worker %d ended", id)
			if err != nil {
				errs[id] = err
			}
		}(i)
	}

	for {
		for id := range playlists.Items {
			// Items is already array of pointers
			p := playlists.Items[id]
			pch <- p
		}

		if playlists.NextPageToken == "" {
			break
		} else {
			playlists, err = state.youtube.Playlists.List([]string{"snippet"}).PageToken(playlists.NextPageToken).Id(b.config.YoutubeSavedPlaylistIds...).MaxResults(50).Do()
			if err != nil {
				close(pch)
				return
			}
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
func (b *backuper) worker_youtube(st *backupState, playlists <-chan *gyoutube.Playlist) (err error) {
	defer st.wg.Done()

	for {
		select {
		case p := <-playlists:
			if p == nil {
				log.Debug().Msgf("backuper_worker_youtube: received nil playlist, exiting")
				return
			}

			log.Debug().Msgf("backuper_worker_youtube: received playlist '%s' with pointer '%p'", p.Snippet.Title, p)
			if err != nil {
				log.Warn().Err(err).Msgf("backuper_worker_youtube: skipping playlist '%s' because error'ed already", p.Snippet.Title)
				continue
			}

			err = b.savePlaylistYoutube(st, p)
			if err != nil {
				// don't exit, try to save other playlists
				log.Error().Err(err).Msgf("backuper_worker_youtube: encountered an error while saving playlist '%s'", p.Snippet.Title)
			}
		case <-st.ctx.Done():
			log.Debug().Msg("backuper_worker_youtube: exiting")
			return
		}
	}
}

func (b *backuper) savePlaylistYoutube(st *backupState, playlist *gyoutube.Playlist) (err error) {
	p, err := b.addYoutubePlaylist(st.bp, playlist)
	if err != nil {
		log.Error().Err(err).Msgf("backuper_worker_youtube: could not create playlist entry for '%s'", playlist.Snippet.Title)
		return
	}

	pageToken := ""
	for {
		call := st.youtube.PlaylistItems.List([]string{"id", "snippet", "contentDetails"}).PlaylistId(playlist.Id).MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		tracks, err := call.Do()
		if err != nil {
			return err
		}

		log.Debug().Msgf("backuper_worker_youtube: got track page for '%s', total %d", playlist.Snippet.Title, tracks.PageInfo.TotalResults)

		pageToken = tracks.NextPageToken

		for _, t := range tracks.Items {
			log.Debug().Msgf("backuper_worker_youtube: playlist '%s', track '%s'", playlist.Snippet.Title, t.Snippet.Title)

			err = b.addYoutubeTrack(st.bp, p, t)
			if err != nil {
				log.Error().Err(err).Msgf("backuper_worker_youtube: could not create track entry for '%s'/'%s'", playlist.Snippet.Title, t.Snippet.Title)
				return err
			}
		}

		if pageToken == "" {
			break
		}
	}

	return nil
}
