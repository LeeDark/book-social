package users

import "time"

const SessionTokenHashSize = 32

// User is the identity exposed to application services after authentication.
// It deliberately does not contain a password or password hash.
type User struct {
	ID        int
	FirstName string
	Login     string
	Email     string
	RoleID    int
}

type Role struct {
	ID      int
	Name    string
	IsAdmin bool
}

// Credentials is a repository result for the password verification boundary.
// Services must keep PasswordHash inside that boundary and never return it.
type Credentials struct {
	User         User
	PasswordHash string
}

type CreateUserParams struct {
	FirstName    string
	Login        string
	Email        string
	PasswordHash string
	RoleID       int
}

type CreateSessionParams struct {
	UserID    int
	TokenHash []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Session struct {
	ID        int
	UserID    int
	TokenHash []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}
