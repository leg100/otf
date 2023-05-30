package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html/paths"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

type (
	// Renderer renders pages and templates
	Renderer interface {
		pageRenderer
		partialRenderer
	}

	// pageRenderer renders html pages.
	pageRenderer interface {
		// Render an html page pageRenderer renders an html page using the named
		// template. The implementation is expected, in the event of an error,
		// to write the error to the response along with a 5xx response code.
		Render(name string, w http.ResponseWriter, page any)
	}

	// partialRenderer renders html partials.
	partialRenderer interface {
		RenderTemplate(name string, w io.Writer, data any) error
		Error(w http.ResponseWriter, err string, code int)
	}

	renderer struct {
		partialRenderer
	}
)

// NewRenderer constructs a renderer. If developer mode is enabled then
// templates are loaded from disk every time a template is rendered.
func NewRenderer(devMode bool) (*renderer, error) {
	var tr partialRenderer
	if devMode {
		tr = &devRenderer{}
	} else {
		embedded, err := newEmbeddedRenderer()
		if err != nil {
			return nil, err
		}
		tr = embedded
	}
	return &renderer{tr}, nil
}

// Render the page. Note this must be the last thing called in a handler because
// it writes an HTTP5xx to the response if there is an error.
func (r *renderer) Render(name string, w http.ResponseWriter, page any) {
	// purge flash messages from cookie store prior to rendering template
	purgeFlashes(w)
	if err := r.RenderTemplate(name, w, page); err != nil {
		r.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
func newTemplateCache(templates fs.FS, buster *cacheBuster, devMode bool) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(templates, contentTemplatesGlob)
	if err != nil {
		return nil, err
	}

	// template functions
	funcs := sprig.HtmlFuncMap()
	// func to append hash to asset links
	funcs["addHash"] = buster.Path
	funcs["version"] = func() string { return internal.Version }
	funcs["trimHTML"] = func(tmpl template.HTML) template.HTML { return template.HTML(strings.TrimSpace(string(tmpl))) }
	funcs["mergeQuery"] = mergeQuery
	funcs["selected"] = selected
	funcs["checked"] = checked
	funcs["disabled"] = disabled
	funcs["insufficient"] = insufficient
	funcs["devMode"] = func() bool { return devMode }
	// make path helpers available to templates
	for k, v := range paths.FuncMap() {
		funcs[k] = v
	}

	for _, page := range pages {
		name := filepath.Base(page)

		template, err := template.New(name).Funcs(funcs).ParseFS(templates,
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
