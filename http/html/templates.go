package html

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
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
	sessions *scs.SessionManager

	// provide access to routes
	router *mux.Router
}

func (f *templateDataFactory) newTemplateData(r *http.Request, content interface{}) templateData {
	return templateData{
		Content:  content,
		router:   f.router,
		sessions: f.sessions,
		request:  r,
	}
}

type templateData struct {
	// Sidebar menu
	Sidebar *sidebar

	// Content is specific to the content being embedded within the layout.
	Content interface{}

	router *mux.Router

	request *http.Request

	sessions *scs.SessionManager
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

func (td *templateData) routeVars() map[string]string {
	return mux.Vars(td.request)
}

// popFlashMessages retrieves all flash messages from the current session. The
// messages are thereafter discarded from the session.
func (td *templateData) popFlashMessages() (msgs []template.HTML) {
	ctx := td.request.Context()
	if msg := td.sessions.PopString(ctx, otf.FlashSessionKey); msg != "" {
		msgs = append(msgs, template.HTML(msg))
	}
	return
}

func (td *templateData) currentUser() string {
	ctx := td.request.Context()
	return td.sessions.GetString(ctx, otf.UsernameSessionKey)
}

func (td *templateData) currentPath() string {
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
