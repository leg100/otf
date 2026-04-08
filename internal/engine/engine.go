// Package engine manages the CLI engine binaries that carry out run operations.
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
	// ErrInvalidVersion is returned when a engine version string is
	// not a semantic version string (major.minor.patch).
	ErrInvalidVersion = errors.New("invalid engine version")
)

func Engines() []*Engine {
	return []*Engine{
		Terraform,
		Tofu,
	}
}

// Engine represents a CLI capable of carrying out infrastructure as code
// operations, e.g. terraform.
type Engine struct {
	Name           string
	DefaultVersion string
	client         Client
}

// Client provides access to the engine's upstream services.
type Client interface {
	// getLatestVersion retrieves the latest available (semantic) version.
	getLatestVersion(context.Context) (string, error)
	// sourceURL returns the URL for retrieving a given version of the engine
	// binary.
	sourceURL(version string) *url.URL
}

func (e *Engine) String() string { return e.Name }

func (e *Engine) Type() string { return "engine" }

func (e *Engine) Set(v string) error {
	return e.set(v)
}

func (e *Engine) MarshalText() ([]byte, error) {
	return []byte(e.Name), nil
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
	return e.Name, nil
}

func (e *Engine) set(v string) error {
	switch v {
	case "terraform":
		*e = *Terraform
	case "tofu":
		*e = *Tofu
	default:
		return fmt.Errorf("no engine found named %s: must be either 'terraform' or 'tofu'", v)
	}
	return nil
}
