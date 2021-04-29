package backup

import (
	"time"

	"github.com/zmb3/spotify"
)

type Repository interface {
	AddBackup(b *Backup) error
	AddPlaylist(b *Backup, p *Playlist) error
	AddTrack(b *Backup, p *Playlist, t *Track) error

	UpdateBackup(b *Backup) error

	// TODO: Add querying
}

func (b *backuper) createBackup(userId string) (bp *Backup, err error) {
	bp = &Backup{
		UserId:  userId,
		Started: time.Now(),
	}

	err = b.repo.AddBackup(bp)
	return
}

func (b *backuper) endBackup(bp *Backup) (err error) {
	bp.Finished = time.Now()

	err = b.repo.UpdateBackup(bp)

	return
}

func (b *backuper) addSpotifyPlaylist(bp *Backup, sp *spotify.SimplePlaylist) (p *Playlist, err error) {
	p = &Playlist{
		SpotifyId: string(sp.ID),
		Name:      sp.Name,
		Created:   time.Now(),
	}

	err = b.repo.AddPlaylist(bp, p)
	return
}

func (b *backuper) addSpotifyTrack(bp *Backup, p *Playlist, st *spotify.PlaylistTrack) (err error) {
	t := &Track{
		SpotifyId:         string(st.Track.ID),
		Name:              st.Track.Name,
		Artist:            formatTrackArtists(st.Track.Artists),
		Album:             st.Track.Album.Name,
		AddedAtToPlaylist: st.AddedAt,
		Created:           time.Now(),
	}

	err = b.repo.AddTrack(bp, p, t)
	return
}

func formatTrackArtists(artists []spotify.SimpleArtist) string {
	var artist string
	lastId := len(artists) - 1
	for id, v := range artists {
		artist += v.Name
		if id != lastId {
			artist += ", "
		}
	}

	return artist
}
