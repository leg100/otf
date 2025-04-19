package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"

	"github.com/leg100/otf/internal/engine"
)

const (
	DefaultVersion        = "1.6.0"
	hashicorpReleasesHost = "releases.hashicorp.com"
)

func init() {
	engine.Register(&Engine{})
}

type Engine struct{}

func (e *Engine) String() string         { return "terraform" }
func (e *Engine) DefaultVersion() string { return DefaultVersion }

func (e *Engine) SourceURL(version string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   hashicorpReleasesHost,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}
}

func (_ *Engine) Type() string { return "engine" }
func (e *Engine) Set(v string) error {
	return engine.SetFlag(e, v)
}

const latestEndpoint = "https://api.releases.hashicorp.com/v1/releases/terraform/latest"

// GetLatestVersion retrieves the latest version string for terraform, following
// semver syntax (e.g. 1.9.0)
//
// TODO: use ctx
func (e *Engine) GetLatestVersion(_ context.Context) (string, error) {
	return getLatestVersion(latestEndpoint)
}

func getLatestVersion(endpoint string) (string, error) {
	// check releases endpoint
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s return non-200 status code: %s", latestEndpoint, resp.Status)
	}
	// decode endpoint response
	var release struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.Version, nil
}
