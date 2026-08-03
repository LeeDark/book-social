ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.24
ARG APP_NAME=book-social
ARG APP_DIR=/app

FROM golang:${GO_VERSION} AS build

ARG APP_NAME
ARG APP_DIR

WORKDIR ${APP_DIR}

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'sqlite postgres' \
    github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/${APP_NAME} \
    ./cmd/web


FROM alpine:${ALPINE_VERSION}

ARG APP_NAME
ARG APP_DIR

LABEL authors="lee"

WORKDIR ${APP_DIR}

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /out/${APP_NAME} ${APP_DIR}/${APP_NAME}
COPY --from=build /go/bin/migrate /usr/local/bin/migrate
COPY internal/web/ ${APP_DIR}/internal/web/
COPY db/sqlite/ ${APP_DIR}/db/sqlite/
COPY db/postgresql/ ${APP_DIR}/db/postgresql/
COPY docker/entrypoint.sh ${APP_DIR}/docker-entrypoint.sh

RUN apk add --no-cache postgresql-client sqlite \
    && chmod +x ${APP_DIR}/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/book-social"]
