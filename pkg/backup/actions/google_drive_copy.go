package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gdrive "google.golang.org/api/drive/v3"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/hoffs/crispy-musicular/pkg/drive"
	"github.com/rs/zerolog/log"
)

type GoogleDriveBackupAction interface {
	Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) error
}

type googleDriveBackupService struct {
	enabled   bool
	dir       string
	auth      auth.Service
	driveAuth drive.Authenticator
}

func NewGoogleDriveBackupAction(conf *config.AppConfig, auth auth.Service) (a GoogleDriveBackupAction, err error) {
	a = &googleDriveBackupService{conf.DriveActionEnabled, conf.DriveDir, auth, drive.NewAuthenticator(conf.DriveId, conf.DriveSecret, conf.DriveCallback)}
	return
}

// this could be a better format, but this is just easier
// and it doesnt take much effort to remap afterwards
type googleDriveBackup struct {
	Backup    *backup.Backup
	Playlists *[]backup.Playlist
	Tracks    *[]backup.Track
}

func (s *googleDriveBackupService) Do(bp *backup.Backup, p *[]backup.Playlist, t *[]backup.Track) (err error) {
	if !s.enabled {
		log.Debug().Msg("google_drive_backup_action: action is not enabled")
		return nil
	}

	st, err := s.auth.GetState()
	if err != nil {
		return
	}

	if st.DriveRefreshToken == "" {
		return errors.New("google_drive_backup_action: drive refresh token is not set")
	}

	drive, err := s.driveAuth.FromRefreshToken(st.DriveRefreshToken)
	if err != nil {
		return
	}

	backup := &googleDriveBackup{bp, p, t}
	data, err := json.Marshal(backup)
	if err != nil {
		log.Error().Err(err).Msg("google_drive_backup_action: failed to marshal json")
		return
	}

	folder, err := s.getOrCreateFolder(drive)
	if err != nil {
		log.Error().Err(err).Msg("google_drive_backup_action: failed to get folder id")
		return
	}

	if folder.Id == "" {
		log.Error().Msg("google_drive_backup_action: got empty folder id")
		return errors.New("google_drive_backup_action: got empty folder id")
	}

	fname := fmt.Sprintf("%s+%s.json", bp.UserId, bp.Started.Format(time.RFC3339))
	file := &gdrive.File{
		Name:        fname,
		Description: "Backup of Spotify playlists",
		MimeType:    "application/json",
		Parents:     []string{folder.Id},
	}
	r := bytes.NewReader(data)
	file, err = drive.Files.Create(file).Media(r).UseContentAsIndexableText(false).Fields("id,name").Do()
	if err != nil {
		log.Error().Err(err).Msg("google_drive_backup_action: failed to upload file")
	}

	log.Debug().Msgf("google_drive_backup_action: uploaded to google drive with id %s and name %s in folder %s", file.Id, file.Name, folder.Name)
	return
}

func (s *googleDriveBackupService) getOrCreateFolder(drive *gdrive.Service) (*gdrive.File, error) {
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and trashed = false", s.dir)
	files, err := drive.Files.List().Q(query).Fields("files/id,files/name").Do()
	if err != nil {
		return nil, err
	}

	if len(files.Files) > 0 {
		return files.Files[0], nil
	}

	folder := &gdrive.File{
		Name:        s.dir,
		Description: "Directory for backups of Spotify playlists",
		MimeType:    "application/vnd.google-apps.folder",
	}

	created, err := drive.Files.Create(folder).Fields("id,name").Do()
	if err != nil {
		return nil, err
	}

	if created.Id == "" {
		return nil, errors.New("google_drive_backup_action: created folder id is empty")
	}

	return created, nil
}
