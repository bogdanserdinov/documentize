package database

import (
	"context"
	"database/sql"
	"documentize/users"

	_ "github.com/lib/pq" // using postgres driver.
	"github.com/zeebo/errs"

	"documentize"
)

var (
	// Error is the default documentize error class.
	Error = errs.Class("db error")
)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	conn *sql.DB
}

// New returns documentize.DB postgresql implementation.
func New(databaseURL string) (documentize.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{conn: conn}, nil
}

// CreateSchema create schema for all tables and databases.
func (db *database) CreateSchema(ctx context.Context) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users(
	    id         BYTEA PRIMARY KEY               NOT NULL,
	    name       VARCHAR                         NOT NULL,
	    email      VARCHAR                         NOT NULL,
	    status     VARCHAR                         NOT NULL,
	    created_at TIMESTAMP WITH TIME ZONE        NOT NULL
	);`

	_, err := db.conn.ExecContext(ctx, createTableQuery)
	return Error.Wrap(err)
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.conn.Close())
}

func (db *database) Users() users.DB {
	return &usersDB{conn: db.conn}
}
