package html

import (
	"html/template"
	"io/fs"
	"net/http"
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
	// Sidebar menu
	Sidebar *sidebar

	// Flash message to render. Optional.
	Flash template.HTML

	// Username of currently logged in user. Empty if user is not logged in.
	CurrentUser string

	// Breadcrumbs to show current page w.r.t site hierarchy
	Breadcrumbs []anchor

	// Content is specific to the content being embedded within the layout.
	Content interface{}
}

type templateDataOption func(td *templateData)

type sidebar struct {
	Title string
	Items []anchor
}

type anchor struct {
	Name string
	Link string
}

func newTemplateData(r *http.Request, sess *sessions, content interface{}) templateData {
	return templateData{
		Content:     content,
		CurrentUser: sess.currentUser(r),
		Flash:       sess.popFlashMessage(r),
	}
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

func withBreadcrumbs(ancestors ...anchor) templateDataOption {
	return func(td *templateData) {
		td.Breadcrumbs = ancestors
	}
}

func withSidebar(title string, items ...anchor) templateDataOption {
	return func(td *templateData) {
		td.Sidebar = &sidebar{Title: title, Items: items}
	}
}
