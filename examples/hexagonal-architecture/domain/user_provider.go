package domain

import (
	"context"
	"strconv"
	"time"
)

type UserProvider struct {
	repo UserRepo
}

func NewUserProvider(repo UserRepo) (*UserProvider, error) {
	ctx := context.Background()
	// we check if we've already populated the DB with Jean
	_, err := repo.FindUserByName(ctx, "Jean")
	if err != nil {
		_, err = repo.CreateUser(ctx, User{ID: 1, Name: "Jean", EmailAddress: "jean@example.com", BirthDate: time.Date(1900, time.November, 15, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			return nil, err
		}
	}

	return &UserProvider{
		repo: repo,
	}, nil
}

func (up *UserProvider) Create(ctx context.Context, u User) (User, error) {
	createdUser, err := up.repo.CreateUser(ctx, u)
	if err != nil {
		return u, err
	}
	return createdUser, nil
}

func (up *UserProvider) Get(ctx context.Context, id string) (User, error) {
	if id == "" {
		return User{}, nil
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return User{}, err
	}

	u, err := up.repo.FindUserByID(ctx, idInt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (up *UserProvider) Delete(ctx context.Context, id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	err = up.repo.DeleteUser(ctx, idInt)
	if err != nil {
		return err
	}
	return nil
}

func (up *UserProvider) Update(ctx context.Context, u User) error {
	err := up.repo.UpdateUser(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func (up UserProvider) List(ctx context.Context, offset, limit int) ([]User, error) {
	users, err := up.repo.ListUsers(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}
