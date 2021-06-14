package backup

import (
	"time"

	"github.com/zmb3/spotify"
	"google.golang.org/api/youtube/v3"
)

type Repository interface {
	AddBackup(b *Backup) error
	AddPlaylist(b *Backup, p *Playlist) error
	AddTrack(b *Backup, p *Playlist, t *Track) error

	AddYoutubePlaylist(b *Backup, p *YoutubePlaylist) error
	AddYoutubeTrack(b *Backup, p *YoutubePlaylist, t *YoutubeTrack) error

	UpdateBackup(b *Backup) error

	GetLastBackup(userId string) (*Backup, error)
	GetBackupPlaylistCount(b *Backup) (int64, error)
	GetBackupTrackCount(b *Backup) (int64, error)
	GetBackupCount(userId string) (count int64, err error)
	GetBackupData(b *Backup) (p *[]Playlist, t *[]Track, yp *[]YoutubePlaylist, yt *[]YoutubeTrack, err error)
}

func (b *backuper) createBackup(userId string) (bp *Backup, err error) {
	bp = &Backup{
		UserId:  userId,
		Started: time.Now(),
	}

	err = b.repo.AddBackup(bp)
	return
}

func (b *backuper) endBackup(bp *Backup, isOk bool) (err error) {
	bp.Finished = time.Now()
	bp.Success = isOk

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

func (b *backuper) addYoutubePlaylist(bp *Backup, sp *youtube.Playlist) (p *YoutubePlaylist, err error) {
	p = &YoutubePlaylist{
		YoutubeId: sp.Id,
		Name:      sp.Snippet.Title,
		Created:   time.Now(),
	}

	err = b.repo.AddYoutubePlaylist(bp, p)
	return
}

func (b *backuper) addYoutubeTrack(bp *Backup, p *YoutubePlaylist, st *youtube.PlaylistItem) (err error) {
	t := &YoutubeTrack{
		YoutubeId:         st.ContentDetails.VideoId,
		Name:              st.Snippet.Title,
		ChannelTitle:      st.Snippet.VideoOwnerChannelTitle,
		AddedAtToPlaylist: st.Snippet.PublishedAt,
		Created:           time.Now(),
	}

	err = b.repo.AddYoutubeTrack(bp, p, t)
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

type BackupStats struct {
	StartedAt     time.Time
	FinishedAt    time.Time
	Successful    bool
	PlaylistCount int64
	TrackCount    int64
	TotalBackups  int64
}

func (b *backuper) GetBackupStats(userId string) (stats *BackupStats, err error) {
	stats = &BackupStats{}
	bp, err := b.repo.GetLastBackup(userId)
	if err != nil {
		return
	}

	stats.StartedAt = bp.Started
	stats.FinishedAt = bp.Finished
	stats.Successful = bp.Success

	stats.PlaylistCount, err = b.repo.GetBackupPlaylistCount(bp)
	if err != nil {
		return
	}

	stats.TrackCount, err = b.repo.GetBackupTrackCount(bp)
	if err != nil {
		return
	}

	stats.TotalBackups, err = b.repo.GetBackupCount(userId)
	return
}
