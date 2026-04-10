package policy

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gofrs/flock"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/sdassow/atomic"
)

var DefaultBinDir = path.Join(os.TempDir(), "otf-policy-engine-bins")

type binaryResolver interface {
	Resolve(context.Context, *PolicySet, io.Writer) (string, error)
}

type engineSource interface {
	DefaultVersion() string
	String() string
	getLatestVersion(context.Context) (string, error)
	sourceURL(version string) *url.URL
}

type cliResolver struct {
	logger  logr.Logger
	destdir string
	lock    *flock.Flock
	client  *http.Client
	sources map[Kind]engineSource
}

func newCLIResolver(logger logr.Logger, destdir string) (*cliResolver, error) {
	if destdir == "" {
		destdir = DefaultBinDir
	}

	lockFile := filepath.Join(destdir, "otf.lock")
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o777); err != nil {
		return nil, fmt.Errorf("creating directory for lock file: %w", err)
	}

	return &cliResolver{
		logger:  logger.WithValues("component", "policy-engine-resolver"),
		destdir: destdir,
		lock:    flock.New(lockFile),
		client:  &http.Client{},
		sources: map[Kind]engineSource{
			SentinelKind: sentinelEngine{},
		},
	}, nil
}

func (r *cliResolver) Resolve(ctx context.Context, set *PolicySet, out io.Writer) (string, error) {
	source, ok := r.sources[set.Kind]
	if !ok {
		return "", fmt.Errorf("policy engine %q is not implemented", set.Kind)
	}

	version := source.DefaultVersion()
	if set.EngineVersion.Latest {
		latest, err := source.getLatestVersion(ctx)
		if err != nil {
			return "", fmt.Errorf("resolving latest %s version: %w", set.Kind, err)
		}
		version = latest
	} else if set.EngineVersion.String() != "" {
		version = set.EngineVersion.String()
	}

	dest := path.Join(r.destdir, source.String(), version, source.String())
	if internal.Exists(dest) {
		return dest, nil
	}

	_, err := r.lock.TryLockContext(ctx, 678*time.Millisecond)
	if err != nil {
		return "", err
	}
	defer func() {
		if unlockErr := r.lock.Unlock(); unlockErr != nil {
			r.logger.Error(unlockErr, "unlocking lockfile", "kind", set.Kind, "version", version)
		}
	}()

	if internal.Exists(dest) {
		return dest, nil
	}

	download := policyEngineDownload{
		Writer:  out,
		version: version,
		src:     source.sourceURL(version).String(),
		dest:    dest,
		binary:  source.String(),
		client:  r.client,
	}
	if err := download.download(ctx); err != nil {
		return "", err
	}
	return dest, nil
}

type policyEngineDownload struct {
	io.Writer
	version   string
	src, dest string
	binary    string
	client    *http.Client
}

func (d *policyEngineDownload) download(ctx context.Context) error {
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

func (d *policyEngineDownload) getZipfile(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.src, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	res, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 HTTP code: %d", res.StatusCode)
	}

	tmp, err := os.CreateTemp("", "policy-engine-download-*")
	if err != nil {
		return "", fmt.Errorf("creating placeholder for download: %w", err)
	}
	defer tmp.Close()

	if d.Writer != nil {
		fmt.Fprintf(d.Writer, "downloading %s, version %s\n", d.binary, d.version)
	}
	if _, err := io.Copy(tmp, res.Body); err != nil {
		return "", fmt.Errorf("copying to disk: %w", err)
	}
	return tmp.Name(), nil
}

func (d *policyEngineDownload) unzip(zipfile string) error {
	zr, err := zip.OpenReader(zipfile)
	if err != nil {
		return fmt.Errorf("opening archive: %s: %w", zipfile, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name != d.binary {
			continue
		}
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
	return fmt.Errorf("binary not found")
}

const (
	defaultSentinelVersion = "0.40.0"
	sentinelLatestEndpoint = "https://api.releases.hashicorp.com/v1/releases/sentinel/latest"
)

type sentinelEngine struct{}

func (sentinelEngine) String() string         { return "sentinel" }
func (sentinelEngine) DefaultVersion() string { return defaultSentinelVersion }

func (sentinelEngine) sourceURL(version string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "releases.hashicorp.com",
		Path: path.Join(
			"sentinel",
			version,
			fmt.Sprintf("sentinel_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}
}

func (sentinelEngine) getLatestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sentinelLatestEndpoint, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned non-200 status code: %s", sentinelLatestEndpoint, res.Status)
	}
	var release struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.Version, nil
}
