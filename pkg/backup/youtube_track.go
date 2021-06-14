package backup

import "time"

type YoutubeTrack struct {
	Id                uint64
	YoutubeId         string
	Name              string
	ChannelTitle      string
	AddedAtToPlaylist string
	Created           time.Time

	// required when json format backup is written to create correlation
	PlaylistId uint64
}
