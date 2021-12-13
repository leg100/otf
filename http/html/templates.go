package html

import (
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/Masterminds/sprig"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
)

type templateData struct {
	// Flash message to render. Optional.
	Flash template.HTML

	// Content is specific to the content being embedded within the layout.
	Content interface{}
}

// newTemplateCache populates a cache of templates: the pattern is used as a
// glob to lookup templates in fsys, and each template is combined with the
// layout template. The cache is keyed according to the basename of the
// template.
func newTemplateCache(templates fs.FS, static *cacheBuster) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(templates, contentTemplatesGlob)
	if err != nil {
		return nil, err
	}

	functions := sprig.GenericFuncMap()
	functions["addHash"] = static.Path

	for _, page := range pages {
		name := filepath.Base(page)

		template, err := template.New(name).Funcs(functions).ParseFS(templates, page, layoutTemplatePath)
		if err != nil {
			return nil, err
		}

		cache[name] = template
	}

	return cache, nil
}
