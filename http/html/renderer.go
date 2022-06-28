package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/leg100/otf"
)

const (
	// Paths to static assets relative to the templates filesystem. For use with
	// the newTemplateCache function below.
	layoutTemplatePath   = "static/templates/layout.tmpl"
	contentTemplatesGlob = "static/templates/content/*.tmpl"
	partialTemplatesGlob = "static/templates/partials/*.tmpl"
)

func init() {
}

// renderer is capable of locating and rendering a template.
type renderer interface {
	renderTemplate(name string, w io.Writer, data any) error
}

// embeddedRenderer renders templates embedded in the go bin. Uses cache for
// performance.
type embeddedRenderer struct {
	cache map[string]*template.Template
}

// devRenderer renders templates located on disk. The cache is rebuilt every
// time a template is rendered. For development purposes.
type devRenderer struct{}

func newRenderer(devMode bool) (renderer, error) {
	if devMode {
		return &devRenderer{}, nil
	}
	return newEmbeddedRenderer()
}

func newEmbeddedRenderer() (*embeddedRenderer, error) {
	buster := &cacheBuster{embedded}

	cache, err := newTemplateCache(embedded, buster)
	if err != nil {
		return nil, err
	}

	renderer := embeddedRenderer{
		cache: cache,
	}

	return &renderer, nil
}

func (r *embeddedRenderer) renderTemplate(name string, w io.Writer, data any) error {
	return renderTemplateFromCache(r.cache, name, w, data)
}

func (r *devRenderer) renderTemplate(name string, w io.Writer, data any) error {
	buster := &cacheBuster{localDisk}

	cache, err := newTemplateCache(localDisk, buster)
	if err != nil {
		return err
	}

	return renderTemplateFromCache(cache, name, w, data)
}

func renderTemplateFromCache(cache map[string]*template.Template, name string, w io.Writer, data any) error {
	tmpl, ok := cache[name]
	if !ok {
		return fmt.Errorf("unable to locate template: %s", name)
	}

	// Render tmpl out to a tmp buffer first to prevent error messages from
	// being written to browser
	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, &data); err != nil {
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

	// template functions
	funcs := sprig.HtmlFuncMap()
	// func to append hash to asset links
	funcs["addHash"] = buster.Path
	// make version available to templates
	funcs["version"] = func() string { return otf.Version }
	// make path helpers available to templates
	addHelpersToFuncMap(funcs)

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
