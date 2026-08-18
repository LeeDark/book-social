package users

import "errors"

var (
	ErrLoginTaken         = errors.New("login already taken")
	ErrEmailTaken         = errors.New("email already taken")
	ErrRoleNotFound       = errors.New("role not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrInternal           = errors.New("internal error")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return "validation failed"
	}
	return e.Field + ": " + e.Message
}
