package storage

import "time"

type backup struct {
	Id       uint64
	Started  time.Time
	Finished time.Time
}
