package engine

import (
	"fmt"
	"iter"
	"maps"
	"net/url"

	"github.com/spf13/pflag"
)

type Engine interface {
	pflag.Value

	SourceURL(version string) *url.URL
}

var engines = map[string]Engine{}

func Register(e Engine) {
	engines[e.String()] = e
}

func ListEngines() iter.Seq[Engine] {
	return maps.Values(engines)
}

func SetFlag(e Engine, v string) error {
	engine, ok := engines[v]
	if !ok {
		return fmt.Errorf("no engine found named %s: must be one of %v", v, maps.Keys(engines))
	}
	e = engine
	return nil
}
