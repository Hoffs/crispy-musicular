package backup

import (
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
)

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
