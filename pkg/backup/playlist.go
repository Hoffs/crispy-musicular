package backup

import "time"

type Playlist struct {
	Id        int64
	SpotifyId string
	Name      string
	Created   time.Time
}
