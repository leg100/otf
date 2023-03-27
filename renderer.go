package otf

import (
	"io"
	"net/http"
)

// Renderer renders templated responses to http requests.
type Renderer interface {
	// Render template to http response. Template is provided with access
	// various helpers on the root object (.) and the content can be accessed at
	// .Content.
	Render(path string, w http.ResponseWriter, r *http.Request, content any)
	// RenderTemplate renders template to a writer. No helpers are made
	// available and the content is available on the root object (.) within the
	// template.
	RenderTemplate(path string, w io.Writer, content any) error
}
