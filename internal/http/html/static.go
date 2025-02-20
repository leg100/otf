package html

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/logr"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	// The same files but on the local disk
	localPath = "internal/http/html"
	localDisk = os.DirFS(localPath)
)

type assetsFSKey struct{}

func AssetsFS(ctx context.Context) *CacheBuster {
	if cb, ok := ctx.Value(assetsFSKey{}).(*CacheBuster); ok {
		return cb
	}
	return nil
}

// AddStaticHandler adds a handler to router serving static assets
// (JS, CSS, etc) from within go binary. Enabling developer mode sources files from
// local disk instead and starts a live reload server, which reloads the browser
// whenever static files change.
func AddStaticHandler(logger logr.Logger, r *mux.Router, devMode bool) error {
	var fs http.FileSystem
	if devMode {
		if err := startLiveReloadServer(logger); err != nil {
			return fmt.Errorf("starting livereload server: %w", err)
		}
		fs = &CacheBuster{localDisk}
	} else {
		fs = &CacheBuster{embedded}
	}

	// Add assets filesystem to context to allow templates to access assets
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), assetsFSKey{}, fs)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r = r.NewRoute().Subrouter()

	// Middleware to add cache control headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Instruct browser to cache static content for a very long time (1
			// year), and rely on the cache buster to insert a hash to each
			// requested URL, ensuring any content change invalidates the cache.
			w.Header().Set("Cache-Control", "max-age=31536000")
			next.ServeHTTP(w, r)
		})
	})
	r.PathPrefix("/static/").Handler(http.FileServer(fs)).Methods("GET")
	return nil
}
