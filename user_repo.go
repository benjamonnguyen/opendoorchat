package app

import (
	"context"
)

type UserRepo interface {
	CreateUser(context.Context, User) Error
	GetUser(ctx context.Context, id string) (User, Error)
	SearchUserByEmail(context.Context, string) ([]User, Error)
}

type User interface {
	GetAttributes() map[string]string
	GetEmail() string
	IsVerified() bool
	Name() string
	Validate() error
}
