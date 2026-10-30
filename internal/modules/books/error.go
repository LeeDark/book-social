package books

import "errors"

var (
	ErrBookNotFound        = errors.New("book not found")
	ErrAuthorNotFound      = errors.New("author not found")
	ErrMissingBookInternal = errors.New("missing-book storage operation failed")
)
