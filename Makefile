VERSION ?= v0.0.1
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Tool paths
BIN_DIR     := ./bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
# Current version installed in the project
GOLANGCI_LINT_VERSION := 2.12.2

# App settings
DB_PATH ?= ./data/book_social_dev.db
MIGRATE ?= migrate
MIGRATIONS_DIR ?= ./db/sqlite/migrations
MIGRATIONS_DATABASE_URL ?= sqlite://$(DB_PATH)
COMPOSE_DEV := docker compose -f compose.yaml -f compose.dev.yaml
COMPOSE_STAGE := docker compose -f compose.yaml -f compose.stage.yaml
COMPOSE_PROD := docker compose -f compose.yaml -f compose.prod.yaml

# --- Help ---

.PHONY: help
## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# --- Development ---

.PHONY: tidy
## tidy: format code and tidy mod files
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy

.PHONY: audit
## audit: run quality control checks (tidy, vet, lint, test)
audit: tidy lint test
	@echo 'Verifying module dependencies...'
	go mod verify

.PHONY: run
## run: run the application locally
run:
	APP_ENV=dev go run ./cmd/web

# --- Build ---

.PHONY: build
## build: build the application binary
build:
	go build -ldflags "\
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Version=$(VERSION)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.Commit=$(COMMIT)' \
		-X 'github.com/LeeDark/book-social/internal/buildinfo.BuildDate=$(DATE)' "\
		 -o $(BIN_DIR)/app ./cmd/web

# --- Test & Lint ---

.PHONY: test
## test: run all tests
test:
	go test -v -race -count=1 ./...

.PHONY: lint
## lint: run golangci-lint
lint: .install-linter
	@echo 'Running linter...'
	$(GOLANGCI_LINT) run ./... --config=./.golangci.yml

.PHONY: lint/fix
## lint/fix: run golangci-lint and fix issues
lint/fix: .install-linter
	@echo 'Running linter with --fix...'
	$(GOLANGCI_LINT) run ./... --config=./.golangci.yml --fix

.PHONY: .install-linter
.install-linter:
	@if [ ! -x "$(GOLANGCI_LINT)" ] || ! "$(GOLANGCI_LINT)" version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Installing golangci-lint v$(GOLANGCI_LINT_VERSION)..."; \
		GOBIN="$(abspath $(BIN_DIR))" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION); \
	fi

# --- Database ---

.PHONY: db/reset
## db/reset: reset the local development database
db/reset:
	DB_PATH=$(DB_PATH) ./db/sqlite/reset-dev-db.sh

.PHONY: db/migrate/up
## db/migrate/up: apply pending database migrations
db/migrate/up:
	$(MIGRATE) -path "$(MIGRATIONS_DIR)" -database "$(MIGRATIONS_DATABASE_URL)" up

.PHONY: db/migrate/down
## db/migrate/down: roll back the latest database migration
db/migrate/down:
	$(MIGRATE) -path "$(MIGRATIONS_DIR)" -database "$(MIGRATIONS_DATABASE_URL)" down 1

.PHONY: db/shell
## db/shell: open the local development database in sqlite3
db/shell:
	sqlite3 $(DB_PATH)

# --- Docker & Compose ---

.PHONY: docker/build
## docker/build: build the production-like Docker image
docker/build:
	docker build --progress=plain -t book-social:dev .

.PHONY: compose/dev/up
## compose/dev/up: start the dev Compose environment with SQLite
compose/dev/up:
	$(COMPOSE_DEV) up --build

.PHONY: compose/dev/down
## compose/dev/down: stop the dev Compose environment and remove volumes
compose/dev/down:
	$(COMPOSE_DEV) down -v

.PHONY: compose/stage/up
## compose/stage/up: start the stage Compose environment with PostgreSQL
compose/stage/up:
	$(COMPOSE_STAGE) up --build

.PHONY: compose/stage/down
## compose/stage/down: stop the stage Compose environment and remove volumes
compose/stage/down:
	$(COMPOSE_STAGE) down -v

.PHONY: compose/prod/up
## compose/prod/up: start the prod Compose environment with PostgreSQL
compose/prod/up:
	$(COMPOSE_PROD) up --build

.PHONY: compose/prod/down
## compose/prod/down: stop the prod Compose environment and remove volumes
compose/prod/down:
	$(COMPOSE_PROD) down -v

# --- Tools ---

.PHONY: templ/generate
## templ/generate: generate Go code from .templ files
templ/generate:
	go tool templ generate
