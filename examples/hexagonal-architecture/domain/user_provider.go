package domain

import (
	"context"
	"strconv"
	"time"
)

type SQLUserProvider struct {
	repo UserRepo
}

func NewSQLUserProvider(repo UserRepo) (*SQLUserProvider, error) {
	ctx := context.Background()
	// we check if we've already populated the DB with Jean
	_, err := repo.FindUserByName(ctx, "Jean")
	if err != nil {
		_, err = repo.CreateUser(ctx, User{ID: 1, Name: "Jean", EmailAddress: "jean@example.com", BirthDate: time.Date(1900, time.November, 15, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			return nil, err
		}
	}

	return &SQLUserProvider{
		repo: repo,
	}, nil
}

func (up *SQLUserProvider) Create(ctx context.Context, u User) (User, error) {
	createdUser, err := up.repo.CreateUser(ctx, u)
	if err != nil {
		return u, err
	}
	return createdUser, nil
}

func (up *SQLUserProvider) Get(ctx context.Context, ent string) (User, error) {
	if ent == "" {
		return User{}, nil
	}

	id, err := strconv.Atoi(ent)
	if err != nil {
		return User{}, err
	}

	u, err := up.repo.FindUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (up *SQLUserProvider) Delete(ctx context.Context, idString string) error {
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}

	err = up.repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (up *SQLUserProvider) Update(ctx context.Context, u User) error {
	err := up.repo.UpdateUser(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (up SQLUserProvider) List(ctx context.Context, offset, limit int) ([]User, error) {
	users, err := up.repo.ListUsers(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}
