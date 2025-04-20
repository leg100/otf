package releases

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sdassow/atomic"
)

// download represents a current download of an engine version
type download struct {
	// for outputting progress updates
	io.Writer

	version   string
	src, dest string
	binary    string
	client    *http.Client
}

func (d *download) download(ctx context.Context) error {
	zipfile, err := d.getZipfile(ctx)
	if err != nil {
		return fmt.Errorf("downloading zipfile from %s: %w", d.src, err)
	}
	defer os.Remove(zipfile)

	if err := os.MkdirAll(filepath.Dir(d.dest), 0o777); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := d.unzip(zipfile); err != nil {
		return fmt.Errorf("unzipping archive: %w", err)
	}

	return nil
}

func (d *download) getZipfile(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", d.src, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("received non-200 HTTP code: %d", res.StatusCode)
	}

	tmp, err := os.CreateTemp("", "engine-download-*")
	if err != nil {
		return "", fmt.Errorf("creating placeholder for download: %w", err)
	}
	defer tmp.Close()

	fmt.Fprintf(d, "downloading %s, version %s\n", d.binary, d.version)

	_, err = io.Copy(tmp, res.Body)
	if err != nil {
		return "", fmt.Errorf("copying to disk: %w", err)
	}

	return tmp.Name(), nil
}

func (d *download) unzip(zipfile string) error {
	zr, err := zip.OpenReader(zipfile)
	if err != nil {
		return fmt.Errorf("opening archive: %s: %w", zipfile, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name == d.binary {
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()
			if err := atomic.WriteFile(d.dest, fr, atomic.DefaultFileMode(0o755)); err != nil {
				return fmt.Errorf("writing binary: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("binary not found")
}
