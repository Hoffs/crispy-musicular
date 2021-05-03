package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/rs/zerolog/log"
)

type JsonBackupAction interface {
	Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) error
}

type jsonBackupService struct {
	dir string
}

func NewJsonBackupAction(directory string) (a JsonBackupAction, err error) {
	a = &jsonBackupService{directory}
	err = os.MkdirAll(directory, os.ModePerm)
	return
}

// this could be a better format, but this is just easier
// and it doesnt take much effort to remap afterwards
type jsonBackup struct {
	Backup    *backup.Backup
	Playlists *[]backup.Playlist
	Tracks    *[]backup.Track
}

func (s *jsonBackupService) Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) (err error) {
	fname := fmt.Sprintf("%s+%s.json", bp.UserId, bp.Started.Format(time.RFC3339))
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
