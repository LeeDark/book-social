package users

import (
	"context"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	FindRoleByName(ctx context.Context, name string) (Role, error)
	FindCredentials(ctx context.Context, identifier string) (Credentials, error)
	FindByID(ctx context.Context, id int) (User, error)
}

type RegistrationRepository interface {
	UserRepository
	WithinTransaction(ctx context.Context, fn func(UserRepository) error) error
}

type SessionRepository interface {
	CreateSession(ctx context.Context, params CreateSessionParams) (Session, error)
	LoadSession(ctx context.Context, tokenHash []byte, now time.Time) (Session, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
}
