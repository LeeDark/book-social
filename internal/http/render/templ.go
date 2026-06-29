package render

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func RenderTempl(w http.ResponseWriter, r *http.Request, status int, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if err := component.Render(r.Context(), w); err != nil {
		return fmt.Errorf("render templ component: %w", err)
	}

	return nil
}
