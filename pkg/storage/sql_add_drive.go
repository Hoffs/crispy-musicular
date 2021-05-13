package storage

var addDriveSql = `
ALTER TABLE auth_state
	ADD COLUMN drive_refresh_token TEXT;

PRAGMA user_version=1;
`
