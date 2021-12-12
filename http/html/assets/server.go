package assets

import (
	"context"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Masterminds/sprig"
)

const (
	// Paths to static assets in relation to the package directory
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	stylesheetDir        = "static/css"
)

// Server provides the means to retrieve html-related assets (templates and
// static files such as CSS).
type Server interface {
	RenderTemplate(ctx context.Context, name string, w io.Writer, data interface{}) error
	GetStaticFS() http.FileSystem
}

type LayoutTemplateOptions struct {
	Title         string
	Stylesheets   []string
	FlashMessages []template.HTML
}

type templateCache struct {
	static *StaticFS
	cache  map[string]*template.Template
}

func newTemplateCache(static *StaticFS) *templateCache {
	return &templateCache{
		static: static,
		cache:  make(map[string]*template.Template),
	}
}

// newTemplateCache populates a cache of templates: the pattern is used as a
// glob to lookup templates in fsys, and each template is combined with the
// layout template. The cache is keyed according to the basename of the
// template.
func (tc *templateCache) populate(pattern, layout string) error {
	pages, err := fs.Glob(tc.static, pattern)
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		template, err := template.New(name).Funcs(sprig.FuncMap()).ParseFS(fsys, page, layout)
		if err != nil {
			return nil, err
		}

		cache[name] = template
	}

	return cache, nil
}
