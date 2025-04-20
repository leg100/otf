package engine

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"net/url"
)

var Default = &Engine{engine: &terraformEngine{}}

type engine interface {
	String() string
	DefaultVersion() string
	GetLatestVersion(context.Context) (string, error)
	SourceURL(version string) *url.URL
}

type Engine struct {
	engine
}

func (*Engine) Type() string { return "engine" }
func (e *Engine) Set(v string) error {
	return e.set(v)
}

func (e *Engine) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Engine) UnmarshalText(text []byte) error {
	return e.set(string(text))
}

func (e *Engine) set(v string) error {
	switch v {
	case "terraform":
		e.engine = &terraformEngine{}
	case "tofu":
		e.engine = &tofuEngine{}
	default:
		return fmt.Errorf("no engine found named %s: must be either 'terraform' or 'tofu'", v)
	}
	return nil
}

var engines = map[string]Engine{}

func ListEngines() iter.Seq[Engine] {
	return maps.Values(engines)
}
