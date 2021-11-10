package storage

import (
	"time"

	"github.com/uptrace/bun"
)

type Stat struct {
	UUID             string    `bun:"default:gen_random_uuid()"`
	Name             string    `bun:"username"`
	UserID           int16     `bun:"user_id"`
	ChatID           int64     `bun:"chat_id"`
	Counter          int       `bun:"counter"`
	AverageTimeSleep float64   `bun:"averagetimesleep"`
	CreatedAt        time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdateAt         bun.NullTime
}

type Date struct {
	UUID      string    `bun:"default:gen_random_uuid()"`
	Name      string    `bun:"username"`
	ChatID    int64     `bun:"chat_id"`
	StartDate string    `bun:"start_date"`
	StopDate  string    `bun:"stop_date"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdateAt  bun.NullTime
}
