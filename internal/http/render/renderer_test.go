package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/testutil"
)

type recordingResponseWriter struct {
	header http.Header
	status int
	body   []byte
}

func (w *recordingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *recordingResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *recordingResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body = append(w.body, body...)
	return len(body), nil
}

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

func TestRendererRenderTemplateFailureDoesNotCommitResponse(t *testing.T) {
	renderer := &Renderer{cache: map[string]*template.Template{
		"broken.tmpl": template.Must(template.New("broken").Parse(`{{define "base"}}{{.Missing}}{{end}}`)),
	}}
	w := &recordingResponseWriter{}

	err := renderer.Render(w, http.StatusOK, "broken.tmpl", struct{}{})
	if err == nil {
		t.Fatal("Render() error = nil, want template execution error")
	}
	if w.status != 0 {
		t.Fatalf("response status = %d, want uncommitted response", w.status)
	}
	if len(w.body) != 0 {
		t.Fatalf("response body = %q, want empty", w.body)
	}
}
