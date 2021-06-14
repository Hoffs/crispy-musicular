package storage

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestCreateNewDatabase(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = createDatabase(conn)

	require.NoError(t, err)
}

func TestMigrateDatabase(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = createDatabase(conn)
	require.NoError(t, err)

	r := &repository{conn}
	r.migrate()

	ver, err := r.getVersion()
	require.NoError(t, err)
	require.Equal(t, maxVer, ver)
}

func TestMigrateDatabaseDoesntDeleteState(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = createDatabase(conn)
	require.NoError(t, err)

	r := &repository{conn}
	_, err = r.db.Exec("INSERT INTO auth_state (refresh_token, user, created) VALUES (?, ?, ?)", "a", "b", time.Now())
	require.NoError(t, err)

	r.migrate()
	stAfter, err := r.GetState()
	require.NoError(t, err)

	require.Equal(t, "a", stAfter.RefreshToken)
	require.Equal(t, "b", stAfter.User)
}

func TestCreateNewRepository(t *testing.T) {
	temp, err := os.CreateTemp("", "temp_db")
	require.NoError(t, err)
	defer os.Remove(temp.Name())

	r, err := NewRepository(temp.Name())
	require.NoError(t, err)

	err = r.Close()
	require.NoError(t, err)

	conn, err := sql.Open("sqlite3", temp.Name())
	rows, err := conn.Query(`
		SELECT
				name
		FROM
				sqlite_master
		WHERE
				type ='table' AND
				name NOT LIKE 'sqlite_%';
	`)

	expectedTables := map[string]bool{
		"auth_state":        false,
		"backups":           false,
		"playlists":         false,
		"tracks":            false,
		"youtube_playlists": false,
		"youtube_tracks":    false,
	}

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		_, ok := expectedTables[name]
		require.True(t, ok, "Table was not expected: %s", name)

		expectedTables[name] = true
	}

	for id, v := range expectedTables {
		require.True(t, v, "Table was not found: %s", id)
	}
}

func TestSetState(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user"})
	require.NoError(t, err)
}

func TestSetStateDrive(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user", DriveRefreshToken: "drive"})
	require.NoError(t, err)
}

func TestGetStateFilled(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user"})
	require.NoError(t, err)

	st, err := r.GetState()
	require.NoError(t, err)
	require.Equal(t, st, auth.State{RefreshToken: "token", User: "user"})
}

func TestGetStateFilledDrive(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user", DriveRefreshToken: "drive"})
	require.NoError(t, err)

	st, err := r.GetState()
	require.NoError(t, err)
	require.Equal(t, st, auth.State{RefreshToken: "token", User: "user", DriveRefreshToken: "drive"})
}

func TestGetStateFilledYoutube(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user", YoutubeRefreshToken: "youtube"})
	require.NoError(t, err)

	st, err := r.GetState()
	require.NoError(t, err)
	require.Equal(t, st, auth.State{RefreshToken: "token", User: "user", YoutubeRefreshToken: "youtube"})
}

func TestGetStateEmpty(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	st, err := r.GetState()
	require.NoError(t, err)
	require.Equal(t, st, auth.State{})
}

func TestClearState(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	err = r.SetState(auth.State{RefreshToken: "token", User: "user"})
	require.NoError(t, err)

	err = r.ClearState()
	require.NoError(t, err)

	st, err := r.GetState()
	require.NoError(t, err)
	require.Equal(t, st, auth.State{})
}
