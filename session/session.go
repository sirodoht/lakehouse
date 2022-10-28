package session

type Session struct {
	ID        int64    `db:"id"`
	UserID    int64    `db:"user_id"`
	TokenHash string `db:"token_hash"`
}
