package releases

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"

	"github.com/leg100/otf/internal"
)

var DefaultEngineBinDir = path.Join(os.TempDir(), "otf-engine-bins")

// downloader downloads engine binaries
type downloader struct {
	destdir string        // destination directory for binaries
	client  *http.Client  // client for downloading from server via http
	mu      chan struct{} // ensures only one download at a time
}

// NewDownloader constructs a terraform downloader, with destdir set as the
// parent directory into which the binaries are downloaded. Pass an empty string
// to use a default.
func NewDownloader(destdir string) *downloader {
	if destdir == "" {
		destdir = DefaultEngineBinDir
	}

	mu := make(chan struct{}, 1)
	mu <- struct{}{}

	return &downloader{
		destdir: destdir,
		client:  &http.Client{},
		mu:      mu,
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
		src:     d.src(version),
		dest:    d.dest(version),
		client:  d.client,
	}).download(ctx)

	d.mu <- struct{}{}

	return d.dest(version), err
}

func (d *downloader) src(version string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   d.host,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}).String()
}

func (d *downloader) dest(version string) string {
	return path.Join(d.destdir, version, "terraform")
}
