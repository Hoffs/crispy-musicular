package storage

import "time"

type playlist struct {
	Id        uint64
	SpotifyId string
	Name      string
	Created   time.Time

	// References
	BackupdId uint64
}
