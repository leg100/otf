package releases

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/engine"
)

var DefaultEngineBinDir = path.Join(os.TempDir(), "otf-engine-bins")

// downloader downloads engine binaries
type downloader struct {
	destdir string        // destination directory for binaries
	client  *http.Client  // client for downloading from server via http
	mu      chan struct{} // ensures only one download at a time
	engine  engine.Engine
}

// NewDownloader constructs a terraform downloader, with destdir set as the
// parent directory into which the binaries are downloaded. Pass an empty string
// to use a default.
func NewDownloader(engine engine.Engine, destdir string) *downloader {
	if destdir == "" {
		destdir = DefaultEngineBinDir
	}

	mu := make(chan struct{}, 1)
	mu <- struct{}{}

	return &downloader{
		destdir: destdir,
		client:  &http.Client{},
		mu:      mu,
		engine:  engine,
	}
}

// Download ensures the given engine version is available on the local
// filesystem and returns its path. Thread-safe: if a Download is in-flight and
// another Download is requested then it'll be made to wait until the former has
// finished.
func (d *downloader) Download(ctx context.Context, version string, w io.Writer) (string, error) {
	if internal.Exists(d.dest(version)) {
		return d.dest(version), nil
	}

	select {
	case <-d.mu:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	err := (&download{
		Writer:  w,
		version: version,
		src:     d.engine.SourceURL(version).String(),
		dest:    d.dest(version),
		binary:  d.engine.String(),
		client:  d.client,
	}).download(ctx)

	d.mu <- struct{}{}

	return d.dest(version), err
}

func (d *downloader) dest(version string) string {
	return path.Join(d.destdir, version, d.engine.String())
}
