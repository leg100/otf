package html

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
)

// Render a template. Wraps the upstream templ handler to carry out additional
// actions every time a template is rendered.
func Render(c templ.Component, w http.ResponseWriter, r *http.Request) {
	// purge flash messages from cookie store prior to rendering template
	purgeFlashes(w)
	// add request to context for templates to access
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	// handle errors
	errHandler := templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Error(w, err.Error(), http.StatusBadRequest)
		})
	})
	templ.Handler(c, errHandler).ServeHTTP(w, r.WithContext(ctx))
}

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}
