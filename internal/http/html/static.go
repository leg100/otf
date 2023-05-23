package html

import (
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
		fs = &cacheBuster{localDisk}
	} else {
		fs = &cacheBuster{embedded}
	}
	r.PathPrefix("/static/").Handler(http.FileServer(fs)).Methods("GET")
	return nil
}
