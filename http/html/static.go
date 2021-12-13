package html

import (
	"embed"
	"net/http"
	"os"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	// The same files but on the local disk
	localDisk = os.DirFS("http/html")
)

// NewStaticServer constructs a cache-busting filesystem of static assets (css,
// javascript, etc). Toggling dev mode enables serving files up from local disk
// instead.
func NewStaticServer(devMode bool) http.FileSystem {
	if devMode {
		return &cacheBuster{localDisk}
	}
	return &cacheBuster{embedded}
}
