package render

import (
	"fmt"
	"html/template"
	"path/filepath"
)

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pagesPattern := "internal/web/templates/pages/*.tmpl"
	pages, err := filepath.Glob(pagesPattern)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no templates found by pattern: %s", pagesPattern)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		//patterns := []string{
		//	"internal/web/templates/base.tmpl",
		//	"internal/web/templates/partials/nav.tmpl",
		//	page,
		//}

		ts, err := template.New(name).ParseFiles("internal/web/templates/base.tmpl")
		//ts, err := template.New(name).ParseFiles(patterns...)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("internal/web/templates/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
