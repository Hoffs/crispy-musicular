package storage

var addYoutubeSql = `
ALTER TABLE auth_state
	ADD COLUMN youtube_refresh_token TEXT;

CREATE TABLE IF NOT EXISTS youtube_playlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	youtube_id TEXT NOT NULL,
	name TEXT NOT NULL,
	created TIMESTAMP NOT NULL,

	backup_id INTEGER NOT NULL,
	FOREIGN KEY(backup_id) REFERENCES backups(id)
);

CREATE TABLE IF NOT EXISTS youtube_tracks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	youtube_id TEXT NOT NULL,
	name TEXT NOT NULL,
	channel_title TEXT NOT NULL,
	added_at_to_playlist TEXT,
	created TIMESTAMP NOT NULL,

	playlist_id INTEGER NOT NULL,
	backup_id INTEGER NOT NULL,

	FOREIGN KEY(playlist_id) REFERENCES youtube_playlists(id),
	FOREIGN KEY(backup_id) REFERENCES backups(id)
);

PRAGMA user_version=2;
`
