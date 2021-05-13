package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

var (
	maxVer     = 1
	migrations = map[int]string{
		1: addDriveSql,
	}
)

func createDatabase(db *sql.DB) error {
	_, err := db.Exec(createDbSql)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize database")
		return err
	}

	log.Info().Msg("initialized database")
	return nil
}

func (r *repository) migrate() (err error) {
	ver, err := r.getVersion()
	if err != nil {
		return
	}

	for ver < maxVer {
		ver, err = r.applyMigration(ver + 1)
		if err != nil {
			return
		}
	}

	return
}

func (r *repository) applyMigration(ver int) (int, error) {
	sql, ok := migrations[ver]
	if !ok {
		return 0, fmt.Errorf("storage: missing migration for version %d", ver)
	}

	oldVer, err := r.getVersion()
	if err != nil {
		return 0, err
	}

	_, err = r.db.Exec(sql)
	if err != nil {
		return 0, err
	}

	ver, err = r.getVersion()
	if err != nil {
		return 0, err
	}

	if oldVer == ver {
		return 0, fmt.Errorf("storage: after migration %d version remains the same", oldVer)
	}

	return ver, nil
}

func (r *repository) getVersion() (ver int, err error) {
	rows := r.db.QueryRow("PRAGMA user_version")

	err = rows.Scan(&ver)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return
	}

	return
}
