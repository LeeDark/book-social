package render

import (
	"fmt"
	"net/http"

	g "maragu.dev/gomponents"
)

func RenderGomponent(w http.ResponseWriter, status int, node g.Node) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if err := node.Render(w); err != nil {
		return fmt.Errorf("render gomponent: %w", err)
	}

	return nil
}
