VERSION ?= v0.0.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags "\
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Version=$(VERSION)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Commit=$(COMMIT)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.BuildDate=$(DATE)' "\
		 -o bin/app ./cmd/web

run:
	go run ./cmd/web

test:
	go test ./...
