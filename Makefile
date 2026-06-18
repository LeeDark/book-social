VERSION ?= v0.0.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
DB_PATH ?= ./data/book_social_dev.db

.PHONY: build run test db-dev-reset db-dev-shell

build:
	go build -ldflags "\
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Version=$(VERSION)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Commit=$(COMMIT)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.BuildDate=$(DATE)' "\
		 -o bin/app ./cmd/web

run:
	go run ./cmd/web

test:
	go test -count=1 ./...

db-dev-reset:
	DB_PATH=$(DB_PATH) ./db/sqlite/reset-dev-db.sh

db-dev-shell:
	sqlite3 $(DB_PATH)
