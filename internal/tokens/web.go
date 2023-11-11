package tokens

import (
	"bytes"
	"net/http"

	"github.com/leg100/otf/internal/http/html"
)

// TokenFlashMessage is a helper for rendering a flash message with an
// authentication token.
func TokenFlashMessage(renderer html.Renderer, w http.ResponseWriter, token []byte) error {
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := renderer.RenderTemplate("token_created.tmpl", buf, string(token)); err != nil {
		return err
	}
	html.FlashSuccess(w, buf.String())
	return nil
}
