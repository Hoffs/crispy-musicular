package storage

import "time"

type playlist struct {
	Id        int64
	SpotifyId string
	Name      string
	Created   time.Time

	// References
	BackupdId int64
}
