package backup

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/hoffs/crispy-musicular/pkg/syncplus"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type backuper struct {
	config  *config.AppConfig
	auth    auth.Service
	repo    Repository
	actions []PostBackupAction
}

type Service interface {
	Backup() (err error)
	RunPeriodically(ctx context.Context)
	GetBackupStats(userId string) (stats *BackupStats, err error)
}

func NewBackuper(c *config.AppConfig, s auth.Service, r Repository, actions ...PostBackupAction) (b Service, err error) {
	if c == nil {
		err = errors.New("backuper: config is nil")
		return
	}

	if s == nil {
		err = errors.New("backuper: config is nil")
		return
	}

	b = &backuper{
		config:  c,
		auth:    s,
		repo:    r,
		actions: actions,
	}
	return
}

type backupState struct {
	ctx     context.Context
	wg      sync.WaitGroup
	spotify spotify.Client
	bp      *Backup
}

func (b *backuper) Backup() (err error) {
	var state backupState

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(b.config.WorkerTimeoutSeconds)*time.Second)

	state.ctx = ctx
	defer cancel()

	st, err := b.auth.GetState()
	if err != nil || !st.IsSet() {
		return
	}

	// There should be no long term issues with this as refresh token doesn't change on subsequent
	// authorizations and probably only changes if auth is revoked for the app or something is reset
	// by Spotify.
	sAuth := spotify.NewAuthenticator("", spotify.ScopePlaylistReadPrivate)
	state.spotify = sAuth.NewClient(&oauth2.Token{RefreshToken: st.RefreshToken})
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
			err := b.worker(&state, pch)
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
			return
		}
	}

	// Close to trigger end of work queue
	close(pch)

	timedOut := syncplus.WaitContext(ctx, &state.wg)
	if timedOut {
		log.Warn().Msg("backuper: workers did not finish in time")
	}

	backupOk := true
	for id, idErr := range errs {
		if idErr != nil {
			backupOk = false
			log.Error().Err(idErr).Msgf("backuper: worker %d errored", id)
		}
	}

	log.Info().Msgf("backuper: finished, is ok: %t", backupOk)
	b.endBackup(state.bp, backupOk)

	if backupOk {
		// run actions on backup
		p, t, err := b.repo.GetBackupData(state.bp)
		if err != nil {
			log.Error().Err(err).Msg("backuper: failed to get backup data")
		} else {
			for _, act := range b.actions {
				err := act.Do(state.bp, p, t)
				if err != nil {
					log.Error().Err(err).Msg("backuper: failed to run post backup action")
				}
			}
		}
	}

	return
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

// should be started as goroutine
func (b *backuper) RunPeriodically(ctx context.Context) {
	log.Info().Msg("backuper_periodic: started")

	for {
		duration := time.Duration(b.config.RunIntervalSeconds) * time.Second
		select {
		case <-time.After(duration):
			err := b.Backup()
			if err != nil {
				log.Error().Err(err).Msg("backuper_periodic: backup finished with errors")
			}
		case <-ctx.Done():
			log.Debug().Msg("backuper_periodic: context finished, stopping backups")
			return
		}
	}
}
