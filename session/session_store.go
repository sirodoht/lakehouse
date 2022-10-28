package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Store interface {
	Insert(context.Context, *Session) (int64, error)
	GetOne(context.Context, string) (*Session, error)
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) Insert(ctx context.Context, d *Session) (int64, error) {
	var id int64
	rows, err := s.db.NamedQuery(`
		INSERT INTO sessions (
			user_id,
			token_hash
		) VALUES (
			:user_id,
			:token_hash
		) RETURNING id`, d)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		rows.Scan(&id)
	}
	return id, nil
}

func (s *SQLStore) GetOne(ctx context.Context, token_hash string) (*Session, error) {
	var sessions []*Session
	err := s.db.SelectContext(
		ctx,
		&sessions,
		`SELECT * FROM sessions WHERE token_hash=$1`,
		token_hash,
	)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no user exists with this username")
	}
	return sessions[0], nil
}
