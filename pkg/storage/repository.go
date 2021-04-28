package storage

import (
	"database/sql"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type Repository interface {
	Close() error

	// Table: auth_state
	GetState() (auth.State, error)
	SetState(auth.State) error
	ClearState() error
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
		log.Error().Err(err).Msg("Failed to initialize database")
		return err
	}

	log.Info().Msg("Initialized database")
	return nil
}

func (r *repository) Close() error {
	return r.db.Close()
}