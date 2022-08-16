package data

import "time"

const (
	VestingsTableName     = "vestings"
	VestingsAccountColumn = "account"
	VestingsSeedColumn    = "seed"
	VestingsStatusColumn  = "status"
	VestingsDateColumn    = "date"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusComplete Status = "complete"
	StatusFailed   Status = "failed"
)

type Vesting struct {
	Id      int64     `db:"id" structs:"-"`
	Account string    `db:"account" structs:"account"`
	Seed    []byte    `db:"seed" structs:"seed"`
	Status  Status    `db:"status" structs:"status"`
	Date    time.Time `db:"date" structs:"date"`
}
