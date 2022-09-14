package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/leg100/otf"
)

const HashicorpReleasesHost = "releases.hashicorp.com"

// TerraformDownloader downloads terraform binaries
type TerraformDownloader struct {
	// server hosting binaries
	host string
	// used to lookup destination path for saving download
	Terraform
	// client for downloading from server via http
	client *http.Client
	// mutex channel
	mu chan struct{}
}

func NewTerraformDownloader() *TerraformDownloader {
	mu := make(chan struct{}, 1)
	mu <- struct{}{}

	return &TerraformDownloader{
		host:      HashicorpReleasesHost,
		Terraform: &TerraformPathFinder{},
		client:    &http.Client{},
		mu:        mu,
	}
}

// Download ensures the given version of terraform is available on the local
// filesystem, returning the path to the binary. If it already exists, download
// will be skipped. Thread-safe: if a download is in-flight and
// another download is requested then it'll made be made to wait until the
// former has finished.
func (d *TerraformDownloader) Download(ctx context.Context, version string, w io.Writer) (string, error) {
	if otf.Exists(d.dest(version)) {
		return d.dest(version), nil
	}

	select {
	case <-d.mu:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	err := (&Download{
		Writer:  w,
		version: version,
		src:     d.src(version),
		dest:    d.dest(version),
		client:  d.client,
	}).Download()

	d.mu <- struct{}{}

	return d.dest(version), err
}

func (d *TerraformDownloader) src(version string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   d.host,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_linux_amd64.zip", version)),
	}).String()
}

func (d *TerraformDownloader) dest(version string) string {
	return d.TerraformPath(version)
}
