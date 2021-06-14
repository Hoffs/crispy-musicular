package backup

import "time"

type YoutubePlaylist struct {
	Id        int64
	YoutubeId string
	Name      string
	Created   time.Time
}
