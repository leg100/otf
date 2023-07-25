package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"

	"github.com/leg100/otf/internal"
)

const HashicorpReleasesHost = "releases.hashicorp.com"

type (
	// terraformDownloader downloads terraform binaries
	terraformDownloader struct {
		*TerraformPathFinder // used to lookup destination path for saving download

		host   string        // server hosting binaries
		client *http.Client  // client for downloading from server via http
		mu     chan struct{} // ensures only one download at a time
	}

	// Downloader downloads a specific version of a binary and returns its path
	Downloader interface {
		Download(ctx context.Context, version string, w io.Writer) (string, error)
	}
)

// NewDownloader constructs a terraform downloader. Pass a path finder to
// customise the location to which the bins are persisted, or pass nil to use
// the default.
func NewDownloader(pathFinder *TerraformPathFinder) *terraformDownloader {
	if pathFinder == nil {
		pathFinder = newTerraformPathFinder(defaultTerraformBinDir)
	}

	mu := make(chan struct{}, 1)
	mu <- struct{}{}

	return &terraformDownloader{
		host:                HashicorpReleasesHost,
		TerraformPathFinder: pathFinder,
		client:              &http.Client{},
		mu:                  mu,
	}
}

// Download ensures the given version of terraform is available on the local
// filesystem and returns its path. Thread-safe: if a Download is in-flight and
// another Download is requested then it'll be made to wait until the
// former has finished.
func (d *terraformDownloader) Download(ctx context.Context, version string, w io.Writer) (string, error) {
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
	}).download()

	d.mu <- struct{}{}

	return d.dest(version), err
}

func (d *terraformDownloader) src(version string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   d.host,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}).String()
}

func (d *terraformDownloader) dest(version string) string {
	return d.TerraformPath(version)
}
