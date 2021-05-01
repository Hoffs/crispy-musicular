package backup

type PostBackupAction interface {
	Do(bp *Backup, p *[]Playlist, t *[]Track) error
}
