package storage

var createDbSql = `
CREATE TABLE IF NOT EXISTS auth_state (
	refresh_token text NOT NULL,
	user text NOT NULL,
	created TIMESTAMP
);
`
