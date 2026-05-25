package render

import (
	"fmt"
	"html/template"
	"net/http"
)

type Renderer struct {
	cache map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	cache, err := newTemplateCache()
	if err != nil {
		return nil, err
	}

	r := &Renderer{
		cache: cache,
	}
	return r, nil
}

func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data any) error {
	ts, ok := r.cache[page]
	if !ok {
		//http.Error(w, "template not found", http.StatusInternalServerError)
		return fmt.Errorf("the template %s does not exist", page)
	}

	w.WriteHeader(status)
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		//http.Error(w, "template error", http.StatusInternalServerError)
		return fmt.Errorf("template error: %w", err)
	}

	return nil
}
