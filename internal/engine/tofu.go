package engine

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v65/github"
)

const defaultTofuVersion = "1.9.0"

type tofu struct{}

func (e *tofu) String() string         { return "tofu" }
func (e *tofu) DefaultVersion() string { return defaultTofuVersion }

func (e *tofu) SourceURL(version string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path: path.Join(
			"opentofu",
			"opentofu",
			"releases",
			"download",
			fmt.Sprintf("v%s", version),
			fmt.Sprintf("tofu_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)),
	}
}

func (e *tofu) GetLatestVersion(ctx context.Context) (string, error) {
	return getLatestTofuVersion(ctx, nil)
}

func getLatestTofuVersion(ctx context.Context, url *string) (string, error) {
	client := github.NewClient(nil)
	if url != nil {
		client.WithEnterpriseURLs(*url, *url)
	}
	latest, _, err := client.Repositories.GetLatestRelease(ctx, "opentofu", "opentofu")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(*latest.Name, "v"), nil
}
