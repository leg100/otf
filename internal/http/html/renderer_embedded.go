package html

import (
	"html/template"
	"io"
	"net/http"
)

// embeddedRenderer renders templates embedded in the go bin. Uses cache for
// performance.
type embeddedRenderer struct {
	cache map[string]*template.Template
}

func newEmbeddedRenderer() (*embeddedRenderer, error) {
	buster := &cacheBuster{embedded}

	cache, err := newTemplateCache(embedded, buster, false)
	if err != nil {
		return nil, err
	}

	renderer := embeddedRenderer{
		cache: cache,
	}

	return &renderer, nil
}

func (r *embeddedRenderer) RenderTemplate(name string, w io.Writer, data any) error {
	return renderTemplateFromCache(r.cache, name, w, data)
}

func (r *embeddedRenderer) Error(w http.ResponseWriter, err string, code int) {
	Error(w, err, code, false)
}
