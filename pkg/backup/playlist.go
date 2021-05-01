package backup

import "time"

type Playlist struct {
	Id        int64 `json:"-"`
	SpotifyId string
	Name      string
	Created   time.Time
}
