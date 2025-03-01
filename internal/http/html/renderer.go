package html

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
)

func Render(c templ.Component, w http.ResponseWriter, r *http.Request, options ...func(*templ.ComponentHandler)) {
	// purge flash messages from cookie store prior to rendering template
	purgeFlashes(w)
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	options = append(options, templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, err.Error())
		})
	}))
	templ.Handler(c, options...).ServeHTTP(w, r.WithContext(ctx))
}

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}
