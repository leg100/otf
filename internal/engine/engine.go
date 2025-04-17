package engine

import (
	"fmt"
	"maps"
	"net/url"

	"github.com/leg100/otf/internal/terraform"
)

type Engine interface {
	Name() string
	SourceURL(version string) *url.URL
}

var engines = map[string]Engine{
	"terraform": &terraform.Engine{},
}

func GetEngine(name string) (Engine, error) {
	engine, ok := engines[name]
	if !ok {
		return nil, fmt.Errorf("no engine found named %s: must be one of %v", name, maps.Keys(engines))
	}
	return engine, nil
}
