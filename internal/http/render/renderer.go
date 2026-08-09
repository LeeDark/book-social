package render

import (
	"bytes"
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

	var body bytes.Buffer
	if err := ts.ExecuteTemplate(&body, "base", data); err != nil {
		//http.Error(w, "template error", http.StatusInternalServerError)
		return fmt.Errorf("template error: %w", err)
	}

	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		return fmt.Errorf("write template: %w", err)
	}
	return nil
}

func (r *Renderer) RenderPartial(w http.ResponseWriter, status int, page string, partial string, data any) error {
	ts, ok := r.cache[page]
	if !ok {
		return fmt.Errorf("the template %s does not exist", page)
	}

	var body bytes.Buffer
	if err := ts.ExecuteTemplate(&body, partial, data); err != nil {
		return fmt.Errorf("template error: %w", err)
	}

	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		return fmt.Errorf("write template: %w", err)
	}
	return nil
}
