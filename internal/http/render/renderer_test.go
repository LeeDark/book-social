package render

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestNewRendererLoadsTemplates(t *testing.T) {
	testutil.ChdirProjectRoot(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}
	if renderer == nil {
		t.Fatal("NewRenderer() returned nil renderer")
	}
}

func TestRendererRenderKnownPage(t *testing.T) {
	testutil.ChdirProjectRoot(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	rec := httptest.NewRecorder()
	data := struct {
		view.Page
		Books []any
	}{
		Page: view.Page{Title: "Books"},
	}

	err = renderer.Render(rec, http.StatusOK, "catalog.tmpl", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Books") {
		t.Fatalf("body does not contain Books: %q", rec.Body.String())
	}
}

func TestRendererRenderMissingPageReturnsError(t *testing.T) {
	testutil.ChdirProjectRoot(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	rec := httptest.NewRecorder()

	err = renderer.Render(rec, http.StatusOK, "missing.tmpl", nil)
	if err == nil {
		t.Fatal("Render() error = nil, want error")
	}
}
