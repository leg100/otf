package html

import (
	"embed"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	// The same files but on the local disk
	localDisk = os.DirFS("internal/http/html")
)

// AddStaticHandler adds a handler to router serving static assets
// (JS, CSS, etc) from within go binary. Dev mode sources files from
// local disk instead.
func AddStaticHandler(r *mux.Router, devMode bool) {
	var fs http.FileSystem
	if devMode {
		fs = &cacheBuster{localDisk}
	} else {
		fs = &cacheBuster{embedded}
	}
	r.PathPrefix("/static/").Handler(http.FileServer(fs)).Methods("GET")
}
