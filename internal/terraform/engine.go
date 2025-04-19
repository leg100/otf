package terraform

import (
	"fmt"
	"net/url"
	"path"
	"runtime"

	"github.com/leg100/otf/internal/engine"
)

func init() {
	engine.Register(&Engine{})
}

type Engine struct{}

func (e *Engine) String() string { return "terraform" }

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
