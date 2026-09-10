package models

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsActive(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
