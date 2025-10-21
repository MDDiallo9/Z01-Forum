package models

import "time"

type Sessions struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}
