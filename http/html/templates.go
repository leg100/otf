package html

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/gorilla/mux"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

// templateDataFactory produces templateData structs
type templateDataFactory struct {
	// for extracting info from current session
	sessions *sessions

	// provide access to routes
	router *mux.Router
}

func (f *templateDataFactory) newTemplateData(r *http.Request, content interface{}) templateData {
	return templateData{
		Content: content,
		// TODO: make these methods instead, and make sessions a field of
		// templateData
		CurrentUser: f.sessions.currentUser(r),
		Flash:       f.sessions.popFlashMessage(r),
		router:      f.router,
		CurrentPath: r.URL.Path,
		request:     r,
	}
}

type templateData struct {
	// Sidebar menu
	Sidebar *sidebar

	// Flash message to render. Optional.
	Flash template.HTML

	// Username of currently logged in user. Empty if user is not logged in.
	CurrentUser string

	CurrentPath string

	// Breadcrumbs to show current page w.r.t site hierarchy
	Breadcrumbs []anchor

	// Content is specific to the content being embedded within the layout.
	Content interface{}

	router *mux.Router

	request *http.Request
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

// path constructs a URL path from the named route and pairs of key values for
// the route variables
func (td *templateData) path(name string, pairs ...string) (string, error) {
	u, err := td.router.Get(name).URLPath(pairs...)
	if err != nil {
		return "", err
	}

	return u.Path, nil
}

// routeVars provides access to the requests's route variables
func (td *templateData) routeVars() map[string]string {
	return mux.Vars(td.request)
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
