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
	config *config.AppConfig
	auth   auth.Service
}

type Backuper interface {
	CreateBackup() (err error)
}

func NewBackuper(c *config.AppConfig, s auth.Service) (b Backuper, err error) {
	if c == nil {
		err = errors.New("backuper: config is nil")
		return
	}

	if s == nil {
		err = errors.New("backuper: config is nil")
		return
	}

	b = &backuper{
		config: c,
		auth:   s,
	}
	return
}

func (b *backuper) CreateBackup() (err error) {
	// workgroup
	// This gets called and does following::
	// 1. Preps state:
	// 1.1. Insert a new entry in database and get backup id.
	// 1.2. Create authenticated Spotify client
	// 2. Spin up a couple of worker goroutines that use a channel and wait for "playlists" to arrive
	// 3. Iterate over user playlists and send them to goroutines
	//
	//	Inside goroutines:
	//		Potentially goroutines could also use goroutines for going over big playslists (future feat)
	//		1. Read playlist info, iterate over playlist songs
	//		2. Store songs to database
	//
	// 4. Stop goroutines with special signal.

	// Does this need local state?
	// TODO: Add actual saving of data

	// Create context with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(b.config.WorkerTimeoutSeconds)*time.Second)
	defer cancel()

	st, err := b.auth.GetState()
	if err != nil || !st.IsSet() {
		return
	}

	log.Debug().Msg(st.RefreshToken)
	t := &oauth2.Token{
		RefreshToken: st.RefreshToken,
	}

	// There should be no long term issues with this as refresh token doesn't change on subsequent
	// authorizations and probably only changes if auth is revoked for the app or something is reset
	// by Spotify.
	sp := spotify.NewAuthenticator("", spotify.ScopePlaylistReadPrivate).NewClient(t)
	usr, err := sp.CurrentUser()
	if err != nil {
		log.Error().Err(err).Msg("backuper: failed to get current user, is refresh token invalid?")
		return
	}

	workers := b.config.WorkerCount
	log.Info().Msgf("backuper: starting backup for %s with %d workers", usr.ID, workers)

	// Use either 61 (default pagination is 20) or if worker amount is higher, worker count + 1,
	// so that in best case scenario we prefetch enough data to saturate all workers.
	// Would maybe make sense to add also an upper bound with math.Min, so that the buffer size would
	// be too big.
	bufferSize := int(math.Max(float64(61), float64(workers+1)))
	pch := make(chan *spotify.SimplePlaylist, bufferSize)

	var wg sync.WaitGroup
	for i := uint8(0); i < workers; i++ {
		wg.Add(1)
		go savePlaylist(ctx, &wg, &sp, pch)
	}

	playlists, err := sp.CurrentUsersPlaylists()
	if err != nil {
		log.Error().Err(err).Msg("backuper: couldn't get initial user playlists")
		return
	}

	for {
		log.Debug().Msgf("backuper: got playlist page, offset %d, limit %d, total %d", playlists.Offset, playlists.Limit, playlists.Total)

		for id := range playlists.Playlists {
			p := &playlists.Playlists[id]
			log.Debug().Msgf("backuper: sending %s to worker with pointer %p", p.Name, p)

			// New struct is created, so sending pointer causes no issues even if next page is loaded before
			// previous page has finished.
			pch <- p
		}

		err = sp.NextPage(playlists)
		if err == spotify.ErrNoMorePages {
			break
		}

		if err != nil {
			return
		}
	}

	// Close to trigger end of work queue
	close(pch)

	timedOut := syncplus.WaitContext(ctx, &wg)
	if timedOut {
		log.Warn().Msg("backuper: workers did not finish in time")
	}

	log.Debug().Msg("backuper: finished")
	return
}

// Processes single playlist and saves information
func savePlaylist(ctx context.Context, wg *sync.WaitGroup, c *spotify.Client, playlists <-chan *spotify.SimplePlaylist) {
	defer wg.Done()

	for {
		select {
		case p := <-playlists:
			if p == nil {
				log.Debug().Msgf("backuper_worker: received nil playlist, exiting")
				return
			}

			log.Debug().Msgf("backuper_worker: received playlist %s with pointer %p", p.Name, p)
		case <-ctx.Done():
			log.Debug().Msg("backuper_worker: exiting")
			return
		}
	}
}

// TODO: Add method to run it on the interval (Ticker)
