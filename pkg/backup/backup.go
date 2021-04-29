package backup

import "time"

type Backup struct {
	Id       int64
	UserId   string
	Started  time.Time
	Finished time.Time
}
