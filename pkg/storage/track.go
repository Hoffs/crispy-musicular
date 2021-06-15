package storage

import "time"

type track struct {
	Id                int64
	SpotifyId         string
	Name              string
	Artist            string
	AddedAtToPlaylist string // This might not exist (in Spotify)
	Created           time.Time

	// References
	PlaylistId int64
	BackupId   int64
}
