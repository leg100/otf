package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
)

const (
	MinTerraformVersion = "1.2.0"

	defaultTerraformVersion = "1.6.0"
	hashicorpReleasesHost   = "releases.hashicorp.com"
	latestEndpoint          = "https://api.releases.hashicorp.com/v1/releases/terraform/latest"
)

type terraformEngine struct{}

func (e *terraformEngine) String() string         { return "terraform" }
func (e *terraformEngine) DefaultVersion() string { return defaultTerraformVersion }

func (e *terraformEngine) SourceURL(version string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   hashicorpReleasesHost,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}
}

// GetLatestVersion retrieves the latest version string for terraform, following
// semver syntax (e.g. 1.9.0)
//
// TODO: use ctx
func (e *terraformEngine) GetLatestVersion(_ context.Context) (string, error) {
	return getLatestTerraformVersion(latestEndpoint)
}

func getLatestTerraformVersion(endpoint string) (string, error) {
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
