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
	defaultTerraformVersion = "1.14.9"
	hashicorpReleasesHost   = "releases.hashicorp.com"
	latestEndpoint          = "https://api.releases.hashicorp.com/v1/releases/terraform/latest"
)

func Terraform() *Engine {
	return &Engine{
		Name:           "terraform",
		DefaultVersion: defaultTerraformVersion,
		client: &terraformClient{
			endpoint: latestEndpoint,
		},
	}
}

type terraformClient struct {
	endpoint string
}

func (f *terraformClient) sourceURL(version string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   hashicorpReleasesHost,
		Path: path.Join(
			"terraform",
			version,
			fmt.Sprintf("terraform_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}
}

// getLatestVersion retrieves the latest version string for terraform, following
// semver syntax (e.g. 1.9.0)
func (f *terraformClient) getLatestVersion(ctx context.Context) (string, error) {
	// check releases endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", f.endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
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
