package backup

import "time"

type Backup struct {
	Id       int64
	Started  time.Time
	Finished time.Time
}
