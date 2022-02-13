package app

import (
	"context"
	"database/sql"
	"embed"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/kohkimakimoto/go-sqlexec"
	_ "github.com/mattn/go-sqlite3"
)

// DB is a database client.
// It supports only sqlite3.
type DB struct {
	*sqlx.DB
	// Path is a sqlite3 database file path.
	Path string
}

func NewDB(dbPath string) (*DB, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &DB{
		DB:   db,
		Path: dbPath,
	}, nil
}

func (db *DB) Close() error {
	if db != nil {
		return db.DB.Close()
	}
	return nil
}

func (db *DB) IsNotExist() bool {
	if _, err := os.Stat(db.Path); os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}

func (db *DB) Init(databaseFs embed.FS) error {
	// schema
	b, err := databaseFs.ReadFile("database/schema.sql")
	if err != nil {
		return err
	}
	schema := sqlexec.SourceString(string(b))

	// seed
	b, err = databaseFs.ReadFile("database/seed.sql")
	if err != nil {
		return err
	}
	seed := sqlexec.SourceString(string(b))

	return sqlexec.Exec(db.DB.DB, schema, seed)
}

type QueryBuilder interface {
	ToSql() (string, []interface{}, error)
}

func (db *DB) GetContextByBuilder(ctx context.Context, dest interface{}, builder QueryBuilder) error {
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	if err := db.GetContext(ctx, dest, query, args...); err != nil {
		return err
	}
	return nil
}

func (db *DB) SelectContextByBuilder(ctx context.Context, dest interface{}, builder QueryBuilder) error {
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	if err := db.SelectContext(ctx, dest, query, args...); err != nil {
		return err
	}
	return nil
}

func (db *DB) CountContextByBuilder(ctx context.Context, builder QueryBuilder) (uint64, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	var count uint64
	if err := db.GetContext(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (db *DB) ExecContextByBuilder(ctx context.Context, builder QueryBuilder) (sql.Result, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func IsErrNoRows(err error) bool {
	return err == sql.ErrNoRows
}
