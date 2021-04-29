package storage

import "time"

type track struct {
	Id                uint64
	SpotifyId         string
	Name              string
	Artist            string
	AddedAtToPlaylist string // This might not exist (in Spotify)
	Created           time.Time

	// References
	PlaylistId uint64
	BackupId   uint64
}
