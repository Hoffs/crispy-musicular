package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/rs/zerolog/log"
)

type JsonBackupAction interface {
	Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) error
	DoYoutube(bp *backup.Backup, p *[]backup.YoutubePlaylist, t *[]backup.YoutubeTrack) (err error)
}

type jsonBackupService struct {
	enabled bool
	dir     string
}

func NewJsonBackupAction(conf *config.AppConfig) (JsonBackupAction, error) {
	act := &jsonBackupService{conf.JsonActionEnabled, conf.JsonDir}
	if act.enabled {
		err := os.MkdirAll(act.dir, os.ModePerm)
		return act, err
	} else {
		return act, nil
	}
}

// this could be a better format, but this is just easier
// and it doesnt take much effort to remap afterwards
type jsonBackup struct {
	Backup    *backup.Backup
	Playlists *[]backup.Playlist
	Tracks    *[]backup.Track
}

func (s *jsonBackupService) Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) (err error) {
	if !s.enabled {
		log.Debug().Msg("json_backup_action: action is not enabled")
		return nil
	}

	fname := fmt.Sprintf("spotify-%s+%s.json", bp.UserId, bp.Started.Format(time.RFC3339))
	fpath := path.Join(s.dir, fname)

	backup := &jsonBackup{bp, p, t}

	data, err := json.Marshal(backup)
	if err != nil {
		log.Error().Err(err).Msg("json_backup_action: failed to marshal json")
		return
	}

	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Error().Err(err).Msgf("json_backup_action: failed to open file at %s", fpath)
		return
	}

	n, err := f.Write(data)
	if err != nil {
		log.Error().Msg("json_backup_action: failed to write")
		return
	}

	log.Debug().Msgf("json_backup_action: wrote %d bytes", n)

	return
}

type youtubeJsonBackup struct {
	Backup    *backup.Backup
	Playlists *[]backup.YoutubePlaylist
	Tracks    *[]backup.YoutubeTrack
}

func (s *jsonBackupService) DoYoutube(bp *backup.Backup, p *[]backup.YoutubePlaylist, t *[]backup.YoutubeTrack) (err error) {
	if !s.enabled {
		log.Debug().Msg("json_backup_action: action is not enabled")
		return nil
	}

	fname := fmt.Sprintf("youtube-%s+%s.json", bp.UserId, bp.Started.Format(time.RFC3339))
	fpath := path.Join(s.dir, fname)

	backup := &youtubeJsonBackup{bp, p, t}

	data, err := json.Marshal(backup)
	if err != nil {
		log.Error().Err(err).Msg("json_backup_action: failed to marshal json")
		return
	}

	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Error().Err(err).Msgf("json_backup_action: failed to open file at %s", fpath)
		return
	}

	n, err := f.Write(data)
	if err != nil {
		log.Error().Msg("json_backup_action: failed to write")
		return
	}

	log.Debug().Msgf("json_backup_action: wrote %d bytes", n)

	return
}
