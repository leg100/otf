package html

import (
	"embed"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/logr"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	AssetsFS *CacheBuster
)

func init() {
	// Serve assets from embedded files in go binary, unless dev mode is
	// enabled, in which case serve files direct from filesystem, to avoid
	// having to rebuild the binary every time a file changes.
	if _, ok := os.LookupEnv("OTF_DEV_MODE"); ok {
		// The working directory differs depending on where go build/test is
		// invoked, so work out the root of the project repo and then join the
		// relative path to the assets.
		wd, err := os.Getwd()
		if err != nil {
			panic(err.Error())
		}
		root := findModuleRoot(wd)
		localPath := filepath.Join(root, "internal/http/html")
		localDisk := os.DirFS(localPath)

		AssetsFS = &CacheBuster{localDisk}
	} else {
		AssetsFS = &CacheBuster{embedded}
	}
}

func findModuleRoot(dir string) (roots string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

// AddStaticHandler adds a handler to router serving static assets
// (JS, CSS, etc) from within go binary. Enabling developer mode sources files from
// local disk instead and starts a live reload server, which reloads the browser
// whenever static files change.
func AddStaticHandler(logger logr.Logger, r *mux.Router) error {
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
	r.PathPrefix("/static/").Handler(http.FileServer(AssetsFS)).Methods("GET")
	return nil
}
