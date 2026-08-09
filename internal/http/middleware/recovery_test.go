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

func TestRecovererLogsStructuredDiagnostics(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(Recoverer(logger))
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
	if entry["msg"] != "http panic recovered" {
		t.Fatalf("log message = %v, want %q", entry["msg"], "http panic recovered")
	}
	if entry["panic"] != "secret panic value" {
		t.Fatalf("panic field = %v, want panic value", entry["panic"])
	}
	if stack, ok := entry["stack"].(string); !ok || !strings.Contains(stack, "TestRecovererLogsStructuredDiagnostics") {
		t.Fatalf("stack field = %v, want test function in stack", entry["stack"])
	}
	if requestID, ok := entry["request_id"].(string); !ok || requestID == "" {
		t.Fatalf("request_id field = %v, want non-empty request ID", entry["request_id"])
	}
}
