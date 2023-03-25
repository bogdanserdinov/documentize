package database

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"documentize/users"
)

var _ users.DB = (*usersDB)(nil)

type usersDB struct {
	conn *sql.DB
}

func (usersDB *usersDB) Create(ctx context.Context, user users.User) error {
	query := `INSERT INTO users(
                  id, 
                  name, 
                  email, 
                  status, 
                  created_at
               ) VALUES (
                   $1,
                   $2,
                   $3,
                   $4,
                   $5
               )`

	_, err := usersDB.conn.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.Status, user.CreatedAt)
	return err
}

func (usersDB *usersDB) Get(ctx context.Context, id uuid.UUID) (users.User, error) {
	query := `SELECT id, name, email, status, created_at
	          FROM users
	          WHERE id = $1`

	var user users.User
	err := usersDB.conn.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Status, &user.CreatedAt)
	return user, err
}

func (usersDB *usersDB) List(ctx context.Context) (_ []users.User, err error) {
	query := `SELECT id, name, email, status, created_at
	          FROM users`

	rows, err := usersDB.conn.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	var list []users.User
	for rows.Next() {
		var user users.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Status, &user.CreatedAt)
		if err != nil {
			return list, err
		}

		list = append(list, user)
	}

	return list, rows.Err()
}

func (usersDB *usersDB) UpdateStatus(ctx context.Context, id uuid.UUID) error {
	result, err := usersDB.conn.ExecContext(ctx, "UPDATE users SET status=$1 WHERE id=$2", "generated", id)
	if err != nil {
		return err
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 {
		return errors.New("user does not exists")
	}

	return nil
}
