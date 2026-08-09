package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func TestRequestLoggerLogsRoutePatternWithoutSource(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		AddSource: true,
	}))

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(RequestLogger(logger))
	router.Get("/books/{bookSlug}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	request := httptest.NewRequest(http.MethodGet, "/books/dune?edition=first", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusCreated)
	}

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v\nlog output: %s", err, logs.String())
	}

	want := map[string]any{
		"method":      http.MethodGet,
		"path":        "/books/dune",
		"query":       "edition=first",
		"status":      float64(http.StatusCreated),
		"bytes":       float64(len("created")),
		"remote_addr": "192.0.2.1:1234",
		"route":       "/books/{bookSlug}",
	}

	for field, value := range want {
		if entry[field] != value {
			t.Errorf("log field %q = %v, want %v", field, entry[field], value)
		}
	}

	if _, ok := entry["duration"]; !ok {
		t.Error("log field \"duration\" is missing")
	}
	if _, ok := entry["request_id"]; !ok {
		t.Error("log field \"request_id\" is missing")
	}
	if _, ok := entry["source"]; ok {
		t.Error("log field \"source\" is present, want omitted for request logs")
	}
}

func TestRequestLoggerLogsRecoveredPanicAsServerError(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(RequestLogger(logger))
	router.Use(chimiddleware.Recoverer)
	router.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("secret panic value")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "secret panic value") {
		t.Fatalf("response leaks panic value: %q", response.Body.String())
	}

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v\nlog output: %s", err, logs.String())
	}
	if entry["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("log status = %v, want %d", entry["status"], http.StatusInternalServerError)
	}
}
