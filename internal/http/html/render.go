package html

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
)

// Render a template. Wraps the upstream templ handler to carry out additional
// actions every time a template is rendered.
func Render(c templ.Component, w http.ResponseWriter, r *http.Request) {
	// add request to context for templates to access
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	// add response to context for templates to access
	ctx = context.WithValue(ctx, responseKey{}, w)
	opts := []func(*templ.ComponentHandler){
		templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// TODO: not sure this is correct; a rendering error more likely
				// means a server side error, so should be an internal server error?
				Error(r, w, err.Error(), http.StatusBadRequest)
			})
		}),
	}
	if r.Header.Get("HX-Target") != "" {
		opts = append(opts, templ.WithFragments(r.Header.Get("Hx-Target")))
	}
	templ.Handler(c, opts...).ServeHTTP(w, r.WithContext(ctx))
}

func RenderSnippet(c templ.Component, w io.Writer, r *http.Request) error {
	// add request to context for templates to access
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	// TODO: is it ok to omit response from context?
	return c.Render(ctx, w)
}

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}

type responseKey struct{}

func ResponseFromContext(ctx context.Context) http.ResponseWriter {
	if w, ok := ctx.Value(responseKey{}).(http.ResponseWriter); ok {
		return w
	}
	return nil
}
