package storage

import (
	"database/sql"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	bp "github.com/hoffs/crispy-musicular/pkg/backup"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type Repository interface {
	Close() error

	// Table: auth_state
	GetState() (auth.State, error)
	SetState(auth.State) error
	ClearState() error

	AddBackup(b *bp.Backup) error
	AddPlaylist(b *bp.Backup, p *bp.Playlist) error
	AddTrack(b *bp.Backup, p *bp.Playlist, t *bp.Track) error

	UpdateBackup(b *bp.Backup) error

	GetLastBackup(userId string) (*bp.Backup, error)
	GetBackupPlaylistCount(b *bp.Backup) (int64, error)
	GetBackupTrackCount(b *bp.Backup) (int64, error)
	GetBackupCount(userId string) (int64, error)
	GetBackupData(b *bp.Backup) (*[]bp.Playlist, *[]bp.Track, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(connString string) (Repository, error) {
	if connString == "" {
		connString = "./data/db.db"
	}

	r := &repository{}

	conn, err := sql.Open("sqlite3", connString)
	if err != nil {
		return r, err
	}

	r.db = conn

	err = createDatabase(conn)
	if err != nil {
		return r, err
	}

	return r, nil
}

func createDatabase(db *sql.DB) error {
	_, err := db.Exec(createDbSql)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize database")
		return err
	}

	log.Info().Msg("initialized database")
	return nil
}

func (r *repository) Close() error {
	return r.db.Close()
}
