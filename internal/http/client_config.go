package http

import (
	"os"
)

const (
	// DefaultBasePath on which the API is served.
	DefaultBasePath = "/api"

	DefaultAddress = "localhost:8080"
)

// Config provides configuration details to the API client.
type Config struct {
	// The address of the otf API.
	Address string
	// The base path on which the API is served.
	BasePath string
	// API token used to access the otf API.
	Token string
	// Insecure skips verification of upstream TLS certs.
	// NOTE: this does not take effect if HTTPClient is non-nil
	Insecure bool
}

// NewConfig constructs a new http client config with defaults.
func NewConfig() Config {
	config := Config{
		Address:  os.Getenv("TFE_ADDRESS"),
		BasePath: DefaultBasePath,
		Token:    os.Getenv("TFE_TOKEN"),
	}
	// Set the default address if none is given.
	if config.Address == "" {
		config.Address = DefaultAddress
	}
	return config
}
