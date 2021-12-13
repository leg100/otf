package html

import (
	"html/template"
	"io"
)

type renderer interface {
	renderTemplate(name string, w io.Writer, data TemplateData) error
}

type embeddedRenderer struct {
	cache map[string]*template.Template

	// filesystem containing static assets
	static *cacheBuster
}

type devRenderer struct {
	// filesystem containing static assets
	static *cacheBuster
}

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

func (r *embeddedRenderer) renderTemplate(name string, w io.Writer, data TemplateData) error {
	return r.cache[name].Execute(w, data)
}

func (r *devRenderer) renderTemplate(name string, w io.Writer, data TemplateData) error {
	static := &cacheBuster{localDisk}

	cache, err := newTemplateCache(localDisk, static)
	if err != nil {
		return err
	}

	return cache[name].Execute(w, data)
}
