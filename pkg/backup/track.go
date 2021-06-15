package backup

import "time"

type Track struct {
	Id                int64
	SpotifyId         string
	Name              string
	Artist            string
	Album             string
	AddedAtToPlaylist string // This might not exist (in Spotify)
	Created           time.Time

	// required when json format backup is written to create correlation
	PlaylistId int64
}
