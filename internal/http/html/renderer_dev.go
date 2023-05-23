package html

import (
	"io"
	"net/http"
)

// devRenderer reads templates from disk. Intended for development purposes.
type devRenderer struct{}

func (r *devRenderer) RenderTemplate(name string, w io.Writer, data any) error {
	buster := &cacheBuster{localDisk}

	cache, err := newTemplateCache(localDisk, buster, true)
	if err != nil {
		return err
	}

	return renderTemplateFromCache(cache, name, w, data)
}

func (r *devRenderer) Error(w http.ResponseWriter, err string, code int) {
	Error(w, err, code, true)
}
