package backup

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
	gyoutube "google.golang.org/api/youtube/v3"
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
	youtube *gyoutube.Service
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

	// Backup database entry is created inside backupSpotify which is not great.
	// That means that Spotify part has to always run first and not continue if
	// it fails.
	backupOk := true
	err = b.backupSpotify(ctx, &state, &st)
	if err != nil {
		backupOk = false
	}

	if backupOk {
		err = b.backupYoutube(ctx, &state, &st)
		if err != nil {
			backupOk = false
		}
	}

	log.Info().Msgf("backuper: finished, is ok: %t", backupOk)
	b.endBackup(state.bp, backupOk)

	if backupOk {
		// run actions on backup
		p, t, yp, yt, err := b.repo.GetBackupData(state.bp)
		if err != nil {
			log.Error().Err(err).Msg("backuper: failed to get backup data")
		} else {
			for _, act := range b.actions {
				err := act.Do(state.bp, p, t)
				if err != nil {
					log.Error().Err(err).Msg("backuper: failed to run post backup action")
				}

				err = act.DoYoutube(state.bp, yp, yt)
				if err != nil {
					log.Error().Err(err).Msg("backuper: failed to run post backup action for youtube")
				}
			}
		}
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
