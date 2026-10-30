package app

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeeDark/book-social/internal/http/cover"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/storage/covers"
	"github.com/LeeDark/book-social/internal/testutil"
)

func TestCoverRouteUsesPublicImagePolicy(t *testing.T) {
	testutil.ChdirProjectRoot(t)
	store, err := covers.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	file, err := store.Put(context.Background(), &data, "a.png", "")
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := render.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := Deps{Logger: logger, Renderer: renderer, CoverHandler: cover.NewHandler(store, logger)}
	application := New(deps, NewHomeHandler(fakeFeaturedBooksProvider{}, renderer, logger), fakeCatalogHandler{})
	for _, method := range []string{"GET", "HEAD", "POST"} {
		w := httptest.NewRecorder()
		application.Router.ServeHTTP(w, httptest.NewRequest(method, file.URL(), nil))
		want := http.StatusOK
		if method == "POST" {
			want = http.StatusMethodNotAllowed
		}
		if w.Code != want || w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Content-Security-Policy") == "" {
			t.Fatalf("%s response = %d, %v", method, w.Code, w.Header())
		}
		if len(w.Result().Cookies()) != 0 {
			t.Fatal("cover route sets session cookies")
		}
	}
}
