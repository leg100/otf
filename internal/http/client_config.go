package http

import (
	"net/http"
	"os"
)

const (
	userAgent        = "go-tfe"
	headerRateLimit  = "X-RateLimit-Limit"
	headerRateReset  = "X-RateLimit-Reset"
	headerAPIVersion = "TFP-API-Version"

	// DefaultBasePath on which the API is served.
	DefaultBasePath = "/api/v2/"
	// PingEndpoint is a no-op API endpoint used to configure the rate limiter
	PingEndpoint = "ping"
)

// RetryLogHook allows a function to run before each retry.
type RetryLogHook func(attemptNum int, resp *http.Response)

// Config provides configuration details to the API client.
type Config struct {
	// The address of the otf API.
	Address string
	// The base path on which the API is served.
	BasePath string
	// API token used to access the otf API.
	Token string
	// Headers that will be added to every request.
	Headers http.Header
	// A custom HTTP client to use.
	HTTPClient *http.Client
	// RetryLogHook is invoked each time a request is retried.
	RetryLogHook RetryLogHook
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
		Headers:  make(http.Header),
	}
	// Set the default address if none is given.
	if config.Address == "" {
		config.Address = DefaultAddress
	}
	// Set the default user agent.
	config.Headers.Set("User-Agent", userAgent)
	return config
}

type rawAPIMetadata struct {
	// APIVersion is the raw API version string reported by the server in the
	// TFP-API-Version response header, or an empty string if that header
	// field was not included in the response.
	APIVersion string
}
