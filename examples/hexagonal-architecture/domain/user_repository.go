package domain

import "context"

type UserRepo interface {
	CreateUser(ctx context.Context, u User) (User, error)
	FindUserByID(ctx context.Context, id int) (User, error)
	FindUserByName(ctx context.Context, name string) (User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, u User) error
	ListUsers(ctx context.Context, offset, limit int) ([]User, error)
}
