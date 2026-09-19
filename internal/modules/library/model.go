package library

import (
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type Item struct {
	ID      int
	Book    books.Book
	AddedAt time.Time
}

type AddItemParams struct {
	UserID  int
	BookID  int
	AddedAt time.Time
}
