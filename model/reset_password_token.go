package model

type Token struct {
	ID        int64
	UserID    int64 `db:user_id`
	Token     string
	IsUsed    bool `db:is_used`
	CreatedAt string `db:"created_at"`
}
