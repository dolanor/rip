package sqluser

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/dolanor/rip"
)

type SQLUserProvider struct {
	db     *sql.DB
	logger *log.Logger
}

func NewSQLUserProvider(db *sql.DB, logger *log.Logger) (*SQLUserProvider, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(id integer primary key autoincrement, name text);")
	if err != nil {
		return nil, err
	}

	return &SQLUserProvider{
		db:     db,
		logger: logger,
	}, nil
}

func (up *SQLUserProvider) Create(ctx context.Context, u User) (User, error) {
	up.logger.Printf("SaveUser: %+v", u)

	res, err := up.db.Exec("INSERT INTO users(name) values(?)", u.Name)
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	u.ID = int(id)
	return u, nil
}

func (up *SQLUserProvider) Get(ctx context.Context, idString string) (User, error) {
	up.logger.Printf("GetUser: %+v", idString)
	if idString == "" {
		return User{}, nil
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return User{}, err
	}

	rows := up.db.QueryRow("SELECT id, name FROM users WHERE id = ?", id)
	if err != nil {
		return User{}, err
	}
	u := User{}
	err = rows.Scan(&u.ID, &u.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, rip.ErrNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (up *SQLUserProvider) Delete(ctx context.Context, idString string) error {
	up.logger.Printf("DeleteUser: %+v", idString)
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}
	_, err = up.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rip.ErrNotFound
		}
		return err
	}

	return nil
}

func (up *SQLUserProvider) Update(ctx context.Context, u User) error {
	up.logger.Printf("UpdateUser: %+v", u.ID)
	_, err := up.db.Exec("UPDATE users SET name = ? WHERE id = ?", u.Name, u.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rip.ErrNotFound
		}
		return err
	}

	return nil
}

func (up SQLUserProvider) List(ctx context.Context, offset, limit int) ([]User, error) {
	up.logger.Printf("ListUser")
	var users []User
	rows, err := up.db.Query("SELECT id, name FROM users LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rip.ErrNotFound
		}
		return nil, err
	}

	for rows.Next() {
		u := User{}
		err = rows.Scan(&u.ID, &u.Name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, rip.ErrNotFound
			}
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
