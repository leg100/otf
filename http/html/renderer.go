package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

// renderer locates and renders a template.
type renderer interface {
	RenderTemplate(name string, w io.Writer, data any) error
}

func newRenderer(devMode bool) (renderer, error) {
	if devMode {
		return &devRenderer{}, nil
	}
	return newEmbeddedRenderer()
}

func renderTemplateFromCache(cache map[string]*template.Template, name string, w io.Writer, data any) error {
	tmpl, ok := cache[name]
	if !ok {
		return fmt.Errorf("unable to locate template: %s", name)
	}

	// Render tmpl out to a tmp buffer first to prevent error messages from
	// being written to browser
	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

// newTemplateCache populates a cache of templates.
func newTemplateCache(templates fs.FS, buster *cacheBuster) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(templates, contentTemplatesGlob)
	if err != nil {
		return nil, err
	}

	// func to append hash to asset links
	FuncMap["addHash"] = buster.Path

	for _, page := range pages {
		name := filepath.Base(page)

		template, err := template.New(name).Funcs(FuncMap).ParseFS(templates,
			layoutTemplatePath,
			partialTemplatesGlob,
			page,
		)
		if err != nil {
			return nil, err
		}

		cache[name] = template
	}

	return cache, nil
}
