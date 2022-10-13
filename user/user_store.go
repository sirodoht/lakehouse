package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store interface {
	Insert(context.Context, *User) (int64, error)
	InsertPage(context.Context, string, string, string) (int64, error)
	GetOne(context.Context, int64) (*User, error)
	Update(context.Context, int64, string, string) (error)
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) Insert(ctx context.Context, d *User) (int64, error) {
	var id int64
	rows, err := s.db.NamedQuery(`
		INSERT INTO users (
			username,
			email,
			created_at,
			updated_at
		) VALUES (
			:username,
			:email,
			:created_at,
			:updated_at
		) RETURNING id`, d)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		rows.Scan(&id)
	}
	return id, nil
}

func (s *SQLStore) InsertPage(ctx context.Context, username string, email string, passwordHash string) (int64, error) {
	var id int64
	timenow := time.Now()
	row := s.db.QueryRow(`
		INSERT INTO users (
			email,
			username,
			password_hash,
			created_at,
			updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		) RETURNING id`,
		email,
		username,
		passwordHash,
		timenow,
		timenow,
	)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SQLStore) Update(ctx context.Context, id int64, field string, value string) error {
	sql := fmt.Sprintf("UPDATE users SET %s=:value WHERE id=:id", field)
	_, err := s.db.NamedExec(sql, map[string]interface{}{
		"field": field,
		"value": value,
		"id":    id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStore) GetOne(ctx context.Context, id int64) (*User, error) {
	var users []*User
	err := s.db.SelectContext(
		ctx,
		&users,
		`SELECT * FROM users WHERE id=$1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return users[0], nil
}
