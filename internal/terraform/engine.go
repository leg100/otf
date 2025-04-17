package terraform

import (
	"fmt"
	"net/url"
	"path"
	"runtime"
)

type Engine struct{}

func (e *Engine) Name() string { return "terraform " }

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
