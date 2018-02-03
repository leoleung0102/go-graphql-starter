package model

import "github.com/satori/go.uuid"

type Token struct {
	ID           uuid.UUID
	UserID       int64 `db:user_id`
	Token        uuid.UUID
	IsExpired    bool `db:is_expired`
	IsUsed       bool `db:is_used`
	CreatedAt    string `db:"created_at"`
}
