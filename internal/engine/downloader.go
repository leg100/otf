package engine

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/gofrs/flock"
	"github.com/leg100/otf/internal"
)

var DefaultBinDir = path.Join(os.TempDir(), "otf-engine-bins")

// downloader downloads engine binaries
type downloader struct {
	destdir string       // destination directory for binaries
	client  *http.Client // client for downloading from server via http
	lock    *flock.Flock // ensures only one download at a time
	engine  engineSource
}

type engineSource interface {
	String() string

	sourceURL(version string) *url.URL
}

// NewDownloader constructs a terraform downloader, with destdir set as the
// parent directory into which the binaries are downloaded. Pass an empty string
// to use a default.
func NewDownloader(engine engineSource, destdir string) *downloader {
	if destdir == "" {
		destdir = DefaultBinDir
	}

	return &downloader{
		destdir: destdir,
		client:  &http.Client{},
		lock:    flock.New("/var/lock/otf.lock"),
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

	_, err := d.lock.TryLockContext(ctx, 678*time.Millisecond)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := d.lock.Unlock(); err != nil {
			// log error
		}
	}()

	// Check if bin exists again, it may have been downloaded whilst caller was
	// blocked on mutex above.
	if internal.Exists(d.dest(version)) {
		return d.dest(version), nil
	}

	err = (&download{
		Writer:  w,
		version: version,
		src:     d.engine.sourceURL(version).String(),
		dest:    d.dest(version),
		binary:  d.engine.String(),
		client:  d.client,
	}).download(ctx)

	return d.dest(version), err
}

func (d *downloader) dest(version string) string {
	return path.Join(d.destdir, version, d.engine.String())
}
