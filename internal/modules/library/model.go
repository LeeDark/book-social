package library

import (
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type ReadingStatus string

const (
	ReadingStatusWantToRead ReadingStatus = "want_to_read"
	ReadingStatusReading    ReadingStatus = "reading"
	ReadingStatusRead       ReadingStatus = "read"
)

type Item struct {
	ID         int
	Book       books.Book
	Status     ReadingStatus
	StartedAt  *time.Time
	FinishedAt *time.Time
	Version    int
	AddedAt    time.Time
}

type AddItemParams struct {
	UserID     int
	BookID     int
	Status     ReadingStatus
	StartedAt  *time.Time
	FinishedAt *time.Time
	Version    int
	AddedAt    time.Time
}

type UpdateStatusParams struct {
	UserID          int
	ItemID          int
	ExpectedVersion int
	Status          ReadingStatus
	StartedAt       *time.Time
	FinishedAt      *time.Time
}
