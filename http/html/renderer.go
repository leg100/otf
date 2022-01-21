package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
)

// renderer is capable of locating and rendering a template.
type renderer interface {
	renderTemplate(name string, w io.Writer, data templateData) error
}

// embeddedRenderer renders templates embedded in the go bin. Uses cache for
// performance.
type embeddedRenderer struct {
	cache map[string]*template.Template
}

// devRenderer renders templates located on disk. No cache is used; ideal for
// development purposes with something like livereload.
type devRenderer struct{}

func newRenderer(devMode bool) (renderer, error) {
	if devMode {
		return &devRenderer{}, nil
	}
	return newEmbeddedRenderer()
}

func newEmbeddedRenderer() (*embeddedRenderer, error) {
	static := &cacheBuster{embedded}

	cache, err := newTemplateCache(embedded, static)
	if err != nil {
		return nil, err
	}

	renderer := embeddedRenderer{
		cache: cache,
	}

	return &renderer, nil
}

func (r *embeddedRenderer) renderTemplate(name string, w io.Writer, data templateData) error {
	return renderTemplateFromCache(r.cache, name, w, data)
}

func (r *devRenderer) renderTemplate(name string, w io.Writer, data templateData) error {
	static := &cacheBuster{localDisk}

	cache, err := newTemplateCache(localDisk, static)
	if err != nil {
		return err
	}

	return renderTemplateFromCache(cache, name, w, data)
}

func renderTemplateFromCache(cache map[string]*template.Template, name string, w io.Writer, data templateData) error {
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
