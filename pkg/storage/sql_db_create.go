package storage

var createDbSql = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS auth_state (
	refresh_token TEXT NOT NULL,
	user TEXT NOT NULL,
	created TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS backups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	started TIMESTAMP NOT NULL,
	finished TIMESTAMP
);

CREATE TABLE IF NOT EXISTS playlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	spotify_id TEXT NOT NULL,
	name TEXT NOT NULL,
	created TIMESTAMP NOT NULL,

	backup_id INTEGER NOT NULL,
	FOREIGN KEY(backup_id) REFERENCES backups(id)
);

CREATE TABLE IF NOT EXISTS tracks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	spotify_id TEXT NOT NULL,
	name TEXT NOT NULL,
	artist TEXT NOT NULL,
	added_at_to_playlist TEXT,
	created TIMESTAMP NOT NULL,

	playlist_id INTEGER NOT NULL,
	backup_id INTEGER NOT NULL,

	FOREIGN KEY(playlist_id) REFERENCES playlists(id),
	FOREIGN KEY(backup_id) REFERENCES backups(id)
);
`
