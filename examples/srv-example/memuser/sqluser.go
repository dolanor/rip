package memuser

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/dolanor/rip"
)

type SQLUserProvider struct {
	db *sql.DB
}

func NewSQLUserProvider(db *sql.DB) (*SQLUserProvider, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(id integer primary key autoincrement, name text);")
	if err != nil {
		return nil, err
	}

	return &SQLUserProvider{
		db: db,
	}, nil
}

func (up *SQLUserProvider) Create(ctx context.Context, u *User) (*User, error) {
	log.Printf("SaveUser: %+v", *u)

	res, err := up.db.Exec("INSERT INTO users(name) values(?)", u.Name)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.ID = int(id)
	return u, nil
}

func (up *SQLUserProvider) Get(ctx context.Context, ent rip.Entity) (*User, error) {
	log.Printf("GetUser: %+v", ent.IDString())
	id, err := strconv.Atoi(ent.IDString())
	if err != nil {
		return nil, err
	}

	rows := up.db.QueryRow("SELECT id, name FROM users WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rip.ErrNotFound
		}
		return nil, err
	}
	u := User{}
	err = rows.Scan(&u.ID, &u.Name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (up *SQLUserProvider) Delete(ctx context.Context, ent rip.Entity) error {
	log.Printf("DeleteUser: %+v", ent.IDString())
	id, err := strconv.Atoi(ent.IDString())
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

func (up *SQLUserProvider) Update(ctx context.Context, u *User) error {
	log.Printf("UpdateUser: %+v", u.IDString())
	_, err := up.db.Exec("UPDATE users SET name = ? WHERE id = ?", u.Name, u.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rip.ErrNotFound
		}
		return err
	}

	return nil
}

func (up SQLUserProvider) ListAll(ctx context.Context) ([]*User, error) {
	log.Printf("ListAllUser")
	var users []*User
	rows, err := up.db.Query("SELECT id, name FROM users")
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
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}
