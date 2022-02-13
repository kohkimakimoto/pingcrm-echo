package app

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Account struct {
	Id        int64      `db:"id"`
	Name      string     `db:"name"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeleteAt  *time.Time `db:"deleted_at"`
}

func GetAccountByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (*Account, error) {
	dest := &Account{}
	if err := db.GetContextByBuilder(ctx, dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func GetAccountById(ctx context.Context, db *DB, id int64) (*Account, error) {
	q := sq.Select("*").From("accounts").Where(sq.Eq{"id": id})
	return GetAccountByBuilder(ctx, db, q)
}
