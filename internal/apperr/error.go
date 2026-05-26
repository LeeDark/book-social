package apperr

import "errors"

var (
	ErrNotFound     = errors.New("page not found")
	ErrInvalidInput = errors.New("invalid input")
)
