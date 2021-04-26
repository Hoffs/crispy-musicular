package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
)

func (r *repository) GetState() (auth.State, error) {
	st := auth.State{}
	rows := r.db.QueryRow("SELECT refresh_token, user FROM auth_state LIMIT 1")

	err := rows.Scan(&st.RefreshToken, &st.User)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}

	if err != nil {
		return st, err
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

	_, err = tx.Exec("INSERT INTO auth_state (refresh_token, user, created) VALUES (?, ?, ?)", st.RefreshToken, st.User, time.Now())
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
