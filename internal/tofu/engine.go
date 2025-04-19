package tofu

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal/engine"
)

const DefaultVersion = "1.9.0"

func init() {
	engine.Register(&Engine{})
}

type Engine struct{}

func (e *Engine) String() string         { return "tofu" }
func (e *Engine) DefaultVersion() string { return DefaultVersion }

func (e *Engine) SourceURL(version string) *url.URL {
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

func (_ *Engine) Type() string { return "engine" }
func (e *Engine) Set(v string) error {
	return engine.SetFlag(e, v)
}

func (e *Engine) GetLatestVersion(ctx context.Context) (string, error) {
	return getLatestVersion(ctx, nil)
}

func getLatestVersion(ctx context.Context, url *string) (string, error) {
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
