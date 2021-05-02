package storage

import (
	"database/sql"
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

func (r *repository) GetLastBackup(userId string) (b *bp.Backup, err error) {
	b = &bp.Backup{UserId: userId}
	result := r.db.QueryRow("SELECT id, started, finished FROM backups WHERE user_id = ? ORDER BY started DESC LIMIT 1", userId)
	var finished sql.NullTime
	err = result.Scan(&b.Id, &b.Started, &finished)
	if finished.Valid {
		b.Finished = finished.Time
	}
	return
}

func (r *repository) GetBackupPlaylistCount(b *bp.Backup) (count int64, err error) {
	result := r.db.QueryRow("SELECT count(*) FROM playlists WHERE backup_id = ?", b.Id)
	err = result.Scan(&count)
	return
}

func (r *repository) GetBackupTrackCount(b *bp.Backup) (count int64, err error) {
	result := r.db.QueryRow("SELECT count(*) FROM tracks WHERE backup_id = ?", b.Id)
	err = result.Scan(&count)
	return
}

func (r *repository) GetBackupCount(userId string) (count int64, err error) {
	result := r.db.QueryRow("SELECT count(*) FROM backups WHERE user_id = ?", userId)
	err = result.Scan(&count)
	return
}

func (r *repository) GetBackupData(b *bp.Backup) (p *[]bp.Playlist, t *[]bp.Track, err error) {
	var lp []bp.Playlist
	var lt []bp.Track
	p = &lp
	t = &lt

	result, err := r.db.Query(
		"SELECT id, spotify_id, name, created FROM playlists WHERE backup_id = ?",
		b.Id)
	if err != nil {
		return
	}

	for result.Next() {
		sp := bp.Playlist{}
		err = result.Scan(&sp.Id, &sp.SpotifyId, &sp.Name, &sp.Created)
		if err != nil {
			return
		}

		lp = append(lp, sp)
	}

	result, err = r.db.Query(
		"SELECT id, spotify_id, name, artist, album, added_at_to_playlist, created, playlist_id FROM tracks WHERE backup_id = ?",
		b.Id)
	if err != nil {
		return
	}

	for result.Next() {
		st := bp.Track{}
		err = result.Scan(&st.Id, &st.SpotifyId, &st.Name, &st.Artist, &st.Album, &st.AddedAtToPlaylist, &st.Created, &st.PlaylistId)
		if err != nil {
			return
		}

		lt = append(lt, st)
	}

	return
}
