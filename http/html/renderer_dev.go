package html

import (
	"io"
	"net/http"
)

// devRenderer reads templates from disk. Intended for development purposes.
type devRenderer struct{}

func (r *devRenderer) RenderTemplate(name string, w io.Writer, data any) error {
	buster := &cacheBuster{localDisk}

	cache, err := newTemplateCache(localDisk, buster)
	if err != nil {
		return err
	}

	return renderTemplateFromCache(cache, name, w, data)
}

func (r *devRenderer) Render(name string, w http.ResponseWriter, page any) {
	if err := r.RenderTemplate(name, w, page); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
	}
}
