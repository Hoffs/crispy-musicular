package storage

import (
	"fmt"

	bp "github.com/hoffs/crispy-musicular/pkg/backup"
)

func (r *repository) AddBackup(b *bp.Backup) (err error) {
	result, err := r.db.Exec("INSERT INTO backups (user_id, started) VALUES (?, ?)", b.UserId, b.Started)
	if err != nil {
		return
	}

	b.Id, err = result.LastInsertId()
	return
}

func (r *repository) AddPlaylist(b *bp.Backup, p *bp.Playlist) (err error) {
	result, err := r.db.Exec(
		"INSERT INTO playlists (spotify_id, name, created, backup_id) VALUES (?, ?, ?, ?)",
		p.SpotifyId,
		p.Name,
		p.Created,
		b.Id)
	if err != nil {
		return
	}

	p.Id, err = result.LastInsertId()
	return
}

func (r *repository) AddTrack(b *bp.Backup, p *bp.Playlist, t *bp.Track) (err error) {
	_, err = r.db.Exec("INSERT INTO tracks (spotify_id, name, artist, album, added_at_to_playlist, created, playlist_id, backup_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		t.SpotifyId,
		t.Name,
		t.Artist,
		t.Album,
		t.AddedAtToPlaylist,
		t.Created,
		p.Id,
		b.Id)
	return
}

func (r *repository) UpdateBackup(b *bp.Backup) (err error) {
	result, err := r.db.Exec("UPDATE backups SET finished = ? WHERE id = ?", b.Finished, b.Id)
	if err != nil {
		return
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 || affected > 1 {
		err = fmt.Errorf("storage: update backup affected %d rows", affected)
	}

	return
}
