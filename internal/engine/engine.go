package engine

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
)

// MinEngineVersion specifies the minimum engine version accepted by OTF.
//
// TODO: This originally applied only to terraform before tofu was added as an
// alternative engine. Tofu's earliest version is 1.6.0, which should really be
// the minimum version if tofu is the selected engine.
const MinEngineVersion = "1.2.0"

var (
	// Default is the default for setting the default engine.
	//
	// NOTE: the actual default engine that has been set by the user should be
	// retrieved via the daemon config.
	Default = Terraform
	// Terraform is the terraform engine
	Terraform = &Engine{engine: &terraform{}}
	// Tofu is the opentofu engine
	Tofu = &Engine{engine: &tofu{}}
	// ErrInvalidVersion is returned when a engine version string is
	// not a semantic version string (major.minor.patch).
	ErrInvalidVersion = errors.New("invalid engine version")
)

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

func (e *Engine) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	return e.set(s)
}

func (e *Engine) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return e.String(), nil
}

func (e *Engine) set(v string) error {
	switch v {
	case "terraform":
		e.engine = &terraform{}
	case "tofu":
		e.engine = &tofu{}
	default:
		return fmt.Errorf("no engine found named %s: must be either 'terraform' or 'tofu'", v)
	}
	return nil
}

func ListEngines() []*Engine {
	return []*Engine{
		{engine: &terraform{}},
		{engine: &tofu{}},
	}
}
