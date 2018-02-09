package model

import "github.com/satori/go.uuid"

type Token struct {
	ID           uuid.UUID
	UserID       uuid.UUID `db:"user_id"`
	Token        string
	IsExpired    bool `db:"is_expired"`
	IsUsed       bool `db:"is_used"`
	CreatedAt    string `db:"created_at"`
}
