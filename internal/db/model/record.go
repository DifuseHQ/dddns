package model

import (
	"time"
)

type Record struct {
	UUID         string    `db:"uuid"`
	Domain       string    `db:"domain"`
	ARecord      string    `db:"a_record"`
	AAAARecord   string    `db:"aaaa_record"`
	CreatedAt    time.Time `db:"created_at"`
	LastUpdateAt time.Time `db:"last_update_at"`
}
