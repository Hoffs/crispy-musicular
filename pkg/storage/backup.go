package storage

import "time"

type backup struct {
	Id       uint64
	UserId   string
	Success  bool
	Started  time.Time
	Finished time.Time
}
