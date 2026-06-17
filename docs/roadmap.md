# Roadmap: Book-Social v1

## About Book-Social v1

## v0.0

- [x] Basic functionality: config, logging, errors, version, Makefile
- [x] Basic HTTP server and routes
- [x] Choose a framework: chi, echo, gin

## v0.1

- [x] Architecture and Infrastructure
  - [x] Basic domain model: User, Catalog, Library
  - [x] Database: SQLite as DB/store for dev, no migrations
  - [x] Basic Dockerfile and Docker Compose
  - [x] Modular monolith, App skeleton
  - [x] Layered Architecture: Handlers & Services & Repositories
- [ ] Frontend
  - [x] Minimal Frontend with Go, MPA, Go Template
  - [ ] HTMX/Templ
- [ ] Basic Home, Catalog routes, handlers and HTML/CSS templates
  - [x] Home, Home to Catalog
  - [ ] Catalog Service
  - [x] Page navigation contract
  - [x] View/Page models
  - [ ] Basic tests

## v0.2

- [ ] Advanced functionality: linters, code quality, CI
- [ ] Database: support SQLite/PostgreSQL for dev/prod
- [ ] Database: the best migration strategy
- [ ] Documentation, MPA Endpoints, OpenAPI
- [ ] Advanced HTTP features: middlewares, graceful shutdown
- [ ] Basic authentication, Sessions and Cookies
- [ ] User Registration, Login, Logout
- [ ] Catalog and User routes, handlers and HTML/CSS templates

## v0.3

- [ ] Database: Data Layer Maturity
  - [ ] Transaction policy
  - [ ] Test fixtures / DB bootstrap
  - [ ] Repository contracts
  - [ ] Admin audit basics
- [ ] Error model, Validation strategy
- [ ] More unit tests
- [ ] Integration tests, try testcontainers
- [ ] Admin panel: Basics

## v0.4

- [ ] User's Library features (to add later)
- [ ] Canonical URL strategy
- [ ] Reading statuses, ratings, notes, shelves/tags
- [ ] Pagination, filters, search
- [ ] Search contract
- [ ] Empty states / UX states
- [ ] Accessibility basics
- [ ] i18n support

## v0.5

- [ ] Catalog features (to add later)
- [ ] Import books from third-pary databases like Goodreads
- [ ] Importer abstraction, background jobs
- [ ] Pagination, filters, search
- [ ] Full-text search in catalog
- [ ] Cover storage

## v0.6

- [ ] User features more deeply, social features step by step: following, feeds, likes
- [ ] Monitoring and tracing with Grafana/Prometheus
