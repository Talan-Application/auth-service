package domain

import "time"

type Code struct {
	ID        int64
	Code      string
	ExpiredAt time.Time
	Receiver  string
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}
