package html

import (
	"io"
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
