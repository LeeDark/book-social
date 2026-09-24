package library

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid library input")
	ErrBookNotFound        = errors.New("library book not found")
	ErrItemAlreadyExists   = errors.New("library item already exists")
	ErrItemNotFound        = errors.New("library item not found")
	ErrItemVersionConflict = errors.New("library item version conflict")
	ErrForbidden           = errors.New("library access forbidden")
	ErrInternal            = errors.New("library internal error")
)
