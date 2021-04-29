package backup

import "time"

type Track struct {
	Id                uint64
	SpotifyId         string
	Name              string
	Artist            string
	Album             string
	AddedAtToPlaylist string // This might not exist (in Spotify)
	Created           time.Time
}
