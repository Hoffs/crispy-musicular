package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
)

func (r *repository) GetState() (auth.State, error) {
	st := auth.State{}
	rows := r.db.QueryRow("SELECT refresh_token, user, drive_refresh_token FROM auth_state LIMIT 1")

	nullDriveRefreshToken := sql.NullString{}
	err := rows.Scan(&st.RefreshToken, &st.User, &nullDriveRefreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}

	if err != nil {
		return st, err
	}

	if nullDriveRefreshToken.Valid {
		st.DriveRefreshToken = nullDriveRefreshToken.String
	}

	return st, nil
}

func (r *repository) SetState(st auth.State) error {
	tx, err := r.db.Begin()
	if err != nil {
		return nil
	}

	_, err = tx.Exec("DELETE FROM auth_state")
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO auth_state (refresh_token, user, drive_refresh_token, created) VALUES (?, ?, ?, ?)", st.RefreshToken, st.User, st.DriveRefreshToken, time.Now())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) ClearState() error {
	_, err := r.db.Exec("DELETE FROM auth_state")
	if err != nil {
		return err
	}
	return nil
}
