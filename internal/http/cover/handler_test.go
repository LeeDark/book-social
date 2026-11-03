package cover

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LeeDark/book-social/internal/storage/covers"
)

func TestImageResponses(t *testing.T) {
	_, handler, data, file, _ := handlerFixture(t)
	tests := []struct {
		name, method, path, header, value string
		status                            int
		body                              []byte
	}{
		{name: "GET", method: "GET", path: file.URL(), status: 200, body: data},
		{name: "HEAD", method: "HEAD", path: file.URL(), status: 200},
		{name: "conditional", method: "GET", path: file.URL(), header: "If-None-Match", value: `"` + file.ChecksumSHA256 + `"`, status: 304},
		{name: "range", method: "GET", path: file.URL(), header: "Range", value: "bytes=0-3", status: 206, body: data[:4]},
		{name: "bad range", method: "GET", path: file.URL(), header: "Range", value: "bytes=999999-", status: 416},
		{name: "POST", method: "POST", path: file.URL(), status: 405},
		{name: "directory", method: "GET", path: "/media/covers/", status: 404},
		{name: "missing", method: "GET", path: "/media/covers/" + strings.Repeat("b", 64) + ".png", status: 404},
		{name: "traversal", method: "GET", path: "/media/covers/../secret.png", status: 404},
		{name: "encoded traversal", method: "GET", path: "/media/covers/%2e%2e%2fsecret.png", status: 404},
		{name: "temporary", method: "GET", path: "/media/covers/.upload-private", status: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.header != "" {
				r.Header.Set(tt.header, tt.value)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", w.Code, tt.status, w.Body.String())
			}
			if tt.status < 400 {
				if !strings.Contains(w.Header().Get("Cache-Control"), "immutable") ||
					w.Header().Get("ETag") != `"`+file.ChecksumSHA256+`"` {
					t.Fatalf("cache headers = %v", w.Header())
				}
				if tt.status != 304 && w.Header().Get("Content-Type") != "image/png" {
					t.Fatalf("Content-Type = %q", w.Header().Get("Content-Type"))
				}
				if !bytes.Equal(w.Body.Bytes(), tt.body) {
					t.Fatalf("body = %q, want %q", w.Body.Bytes(), tt.body)
				}
			} else if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("ETag") != "" {
				t.Fatalf("error cache headers = %v", w.Header())
			}
			if len(w.Result().Cookies()) != 0 {
				t.Fatal("image response sets a cookie")
			}
		})
	}
}

func TestDamagedCoverIsNotServed(t *testing.T) {
	_, handler, _, file, directory := handlerFixture(t)
	if err := os.WriteFile(filepath.Join(directory, file.Key), []byte("damaged"), 0640); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", file.URL(), nil))
	if w.Code != http.StatusNotFound || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("damaged file response = %d, %v", w.Code, w.Header())
	}
}

func TestReadFailureIsGeneric(t *testing.T) {
	store, handler, _, file, _ := handlerFixture(t)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", file.URL(), nil))
	if w.Code != 500 || w.Body.String() != "internal server error\n" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response = %d, %q, %v", w.Code, w.Body.String(), w.Header())
	}
}

func handlerFixture(t *testing.T) (*covers.Store, *Handler, []byte, covers.File, string) {
	t.Helper()
	directory := t.TempDir()
	store, err := covers.New(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	file, err := store.Put(context.Background(), bytes.NewReader(buffer.Bytes()), "a.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".upload-private"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return store, NewHandler(store, logger), buffer.Bytes(), file, directory
}
