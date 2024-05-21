package sqladapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/dolanor/rip/examples/hexagonal-architecture/domain"
)

// ErrorField indicates which field creates the error
type ErrorField string

func (e ErrorField) Error() string {
	return "user field " + string(e) + " in error"
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) (*userRepo, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id integer primary key autoincrement,
		name text,
		email text,
		birth_date date
	);`)
	if err != nil {
		return nil, err
	}
	return &userRepo{
		db: db,
	}, nil
}

func (up *userRepo) CreateUser(ctx context.Context, u domain.User) (domain.User, error) {
	log.Printf("CreateUser: %+v", u)

	res, err := up.db.ExecContext(ctx, `
	INSERT INTO users(
		name,
		email,
		birth_date
	) values (
		?,
		?,
		?
	)`,
		u.Name,
		u.EmailAddress,
		u.BirthDate,
	)
	if err != nil {
		return domain.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}
	u.ID = int(id)
	return u, nil
}

func (up *userRepo) FindUserByID(ctx context.Context, id int) (domain.User, error) {
	log.Printf("FindUserByID: %+v", id)
	row := up.db.QueryRowContext(ctx, `
		SELECT id, name, email, birth_date
		FROM users
		WHERE id = ?`,
		id,
	)
	if row.Err() != nil {
		return domain.User{}, fmt.Errorf("select user query: %w", row.Err())
	}
	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.EmailAddress, &u.BirthDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Join(domain.ErrAppNotFound, ErrorField("ID"))
		}
		return domain.User{}, fmt.Errorf("row scan: %w", err)
	}
	return u, nil
}

func (e ErrorField) ErrorSourcePointer() string {
	return string(e)
}

func (up *userRepo) FindUserByName(ctx context.Context, name string) (domain.User, error) {
	log.Printf("FindUserByName: %+v", name)
	row := up.db.QueryRowContext(ctx, `
		SELECT id, name, email, birth_date
		FROM users
		WHERE name = ?`,
		name,
	)
	if row.Err() != nil {
		return domain.User{}, fmt.Errorf("select user query: %w", row.Err())
	}

	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.EmailAddress, &u.BirthDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Join(domain.ErrAppNotFound, ErrorField("Name"))
		}
		return domain.User{}, fmt.Errorf("row scan: %w", err)
	}
	return u, nil
}

func (up *userRepo) DeleteUser(ctx context.Context, id int) error {
	log.Printf("DeleteUser: %+v", id)

	_, err := up.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAppNotFound
		}
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}

func (up *userRepo) UpdateUser(ctx context.Context, u domain.User) error {
	log.Printf("UpdateUser: %+v", u.ID)
	_, err := up.db.ExecContext(ctx, `
		UPDATE users
		SET
		name = ?,
		email = ?,
		birth_date = ?

		WHERE id = ?
	`,
		u.Name,
		u.EmailAddress,
		u.BirthDate,

		u.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAppNotFound
		}
		return err
	}

	return nil
}

func (up userRepo) ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error) {
	log.Printf("ListUsers")
	var users []domain.User
	//	rows, err := up.db.QueryContext(ctx, "SELECT id, name FROM users")
	rows, err := up.db.QueryContext(ctx, `
		SELECT id, name, email, birth_date
		FROM users
		LIMIT ?
		OFFSET ?`,
		limit,
		offset,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// we return an empty slice
			return users, nil
		}
		return nil, err
	}

	for rows.Next() {
		var u domain.User
		err = rows.Scan(&u.ID, &u.Name, &u.EmailAddress, &u.BirthDate)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
