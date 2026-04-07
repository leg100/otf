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

// Tofu is the opentofu engine.
var Tofu = &Engine{
	Name:                "tofu",
	DefaultVersion:      defaultTofuVersion,
	GetSourceURL:        tofuSourceURL,
	LatestVersionGetter: &tofuLatestVersionGetter{},
}

func tofuSourceURL(version string) *url.URL {
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

type tofuLatestVersionGetter struct {
	endpoint *string
}

func (t *tofuLatestVersionGetter) Get(ctx context.Context) (string, error) {
	client := github.NewClient(nil)
	if t.endpoint != nil {
		var err error
		client, err = client.WithEnterpriseURLs(*t.endpoint, *t.endpoint)
		if err != nil {
			return "", err
		}
	}
	latest, _, err := client.Repositories.GetLatestRelease(ctx, "opentofu", "opentofu")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(*latest.Name, "v"), nil
}
