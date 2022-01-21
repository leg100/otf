package html

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/mux"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

func init() {
	gob.Register(Flash{})
}

// templateDataFactory produces templateData structs
type templateDataFactory struct {
	// for extracting info from current session
	sessions *scs.SessionManager

	// provide access to routes
	router *mux.Router
}

func (f *templateDataFactory) newTemplateData(r *http.Request, content interface{}) templateData {
	return templateData{
		Content:  content,
		router:   &router{f.router},
		sessions: &sessions{f.sessions},
		request:  r,
	}
}

type templateData struct {
	// Content is specific to the content being embedded within the layout.
	Content interface{}

	router *router

	request *http.Request

	sessions *sessions
}

// Path proxies access to the router's route method
func (td *templateData) Path(name string, pairs ...string) string {
	return td.router.route(name, pairs...)
}

// Relative proxies access to the router's relative method
func (td *templateData) Relative(name string, pairs ...string) string {
	return td.router.relative(td.request, name, pairs...)
}

// Ancestor constructs a URL path for the named route, populating its route
// variables using the current route variables. Therefore the named route must
// be an ancestor of the current route, i.e. the named route's variables must be
// a subset of the current route.
func (td *templateData) Ancestor(name string) (string, error) {
	route := td.router.Get(name)

	if route == nil {
		return "", fmt.Errorf("no such web route exists: %s", name)
	}

	pairs := flattenMap(mux.Vars(td.request))

	u, err := route.URLPath(pairs...)
	if err != nil {
		return "", err
	}

	return u.Path, nil
}

func (td *templateData) Breadcrumbs() (crumbs []Anchor, err error) {
	route := mux.CurrentRoute(td.request)

	crumbs, err = td.makeBreadcrumbs(route, crumbs)
	if err != nil {
		return nil, err
	}

	return crumbs, nil
}

func (td *templateData) makeBreadcrumbs(route *mux.Route, crumbs []Anchor) ([]Anchor, error) {
	link, err := route.URLPath(flattenMap(mux.Vars(td.request))...)
	if err != nil {
		return nil, err
	}
	name := path.Base(link.Path)

	// place parent crumb in front
	crumbs = append([]Anchor{{Name: name, Link: link.Path}}, crumbs...)

	parent, ok := parentLookupTable[route.GetName()]
	if !ok {
		return crumbs, nil
	}

	parentRoute := td.router.Get(parent)
	if parentRoute == nil {
		return nil, fmt.Errorf("no such web route exists: %s", parent)
	}

	return td.makeBreadcrumbs(parentRoute, crumbs)
}

func flattenMap(m map[string]string) (s []string) {
	for k, v := range m {
		s = append(s, k, v)
	}
	return
}

func (td *templateData) RouteVars() map[string]string {
	return mux.Vars(td.request)
}

// PopFlashMessages retrieves all flash messages from the current session. The
// messages are thereafter discarded from the session.
func (td *templateData) PopFlashMessages() []Flash {
	return td.sessions.PopAllFlash(td.request)
}

func (td *templateData) CurrentUser() string {
	return td.sessions.CurrentUser(td.request)
}

func (td *templateData) CurrentPath() string {
	return td.request.URL.Path
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
