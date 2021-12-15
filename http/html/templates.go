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
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

type templateData struct {
	// Page title
	Title string

	// Sidebar menu
	Sidebar *sidebar

	// Flash message to render. Optional.
	Flash template.HTML

	// Indicates whether user is currently authenticated or not
	IsAuthenticated bool

	// Content is specific to the content being embedded within the layout.
	Content interface{}
}

type templateDataOption func(td *templateData)

type sidebar struct {
	Title string
	Items []sidebarItem
}

type sidebarItem struct {
	Name string
	Link string
}

// newTemplateCache populates a cache of templates.
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

		template, err := template.New(name).Funcs(functions).ParseFS(templates,
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

func withSidebar(title string, items ...sidebarItem) templateDataOption {
	return func(td *templateData) {
		td.Sidebar = &sidebar{Title: title, Items: items}
	}
}
