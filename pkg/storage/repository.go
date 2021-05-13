package storage

import (
	"database/sql"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	bp "github.com/hoffs/crispy-musicular/pkg/backup"
	_ "github.com/mattn/go-sqlite3"
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

	// If multiple writes happen at same time sqlite might be "locked"
	// this and max open conns should help with that (https://github.com/mattn/go-sqlite3/issues/274)
	opts := "?cache=shared&_journal=WAL"

	conn, err := sql.Open("sqlite3", connString+opts)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)
	r.db = conn

	err = createDatabase(conn)
	if err != nil {
		return nil, err
	}

	err = r.migrate()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *repository) Close() error {
	return r.db.Close()
}
