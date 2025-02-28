package templ

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/http/html"
)

func Render(c templ.Component, w http.ResponseWriter, r *http.Request, options ...func(*templ.ComponentHandler)) {
	html.PopFlashes(r)
	templ.Handler(c, options...).ServeHTTP(w, r)
}
