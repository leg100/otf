package http

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var _ ClientFactory = (*Config)(nil)

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
	// The address of the Terraform Enterprise API.
	Address string
	// The base path on which the API is served.
	BasePath string
	// API token used to access the Terraform Enterprise API.
	Token string
	// Headers that will be added to every request.
	Headers http.Header
	// A custom HTTP client to use.
	HTTPClient *http.Client
	// RetryLogHook is invoked each time a request is retried.
	RetryLogHook RetryLogHook
	// Options for overriding config
	options []ConfigOption
}

type ConfigOption func(*Config) error

// NewConfig constructs a new http client config. Options are only applied when
// NewClient() is called.
func NewConfig(opts ...ConfigOption) (*Config, error) {
	config := &Config{
		Address:    os.Getenv("TFE_ADDRESS"),
		BasePath:   DefaultBasePath,
		Token:      os.Getenv("TFE_TOKEN"),
		Headers:    make(http.Header),
		HTTPClient: cleanhttp.DefaultPooledClient(),
		options:    opts,
	}
	// Set the default address if none is given.
	if config.Address == "" {
		config.Address = DefaultAddress
	}
	// Set the default user agent.
	config.Headers.Set("User-Agent", userAgent)
	return config, nil
}

// NewClient creates a new Terraform Enterprise API client.
func (config *Config) NewClient() (Client, error) {
	// Override config with option args
	for _, o := range config.options {
		o(config)
	}
	var err error
	config.Address, err = SanitizeAddress(config.Address)
	if err != nil {
		return nil, err
	}
	// Parse the address to make sure its a valid URL.
	baseURL, err := url.Parse(config.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}
	baseURL.Path = config.BasePath
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	// This value must be provided by the user.
	if config.Token == "" {
		return nil, fmt.Errorf("missing API token")
	}
	// Create the client.
	client := &client{
		baseURL:      baseURL,
		token:        config.Token,
		headers:      config.Headers,
		retryLogHook: config.RetryLogHook,
	}
	client.http = &retryablehttp.Client{
		Backoff:      client.retryHTTPBackoff,
		CheckRetry:   client.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   config.HTTPClient,
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 400 * time.Millisecond,
		RetryMax:     30,
	}
	meta, err := client.getRawAPIMetadata()
	if err != nil {
		return nil, err
	}
	// Configure the rate limiter.
	client.configureLimiter(meta.RateLimit)
	// Save the API version so we can return it from the RemoteAPIVersion method
	// later.
	client.remoteAPIVersion = meta.APIVersion
	client.ConfigurationVersionService = &configurationVersions{client: client}
	client.EventService = &events{client: client}
	client.OrganizationService = &organizations{client: client}
	//client.Runs = &runs{client: client} client.StateVersionOutputs =
	//&stateVersionOutputs{client: client}
	client.StateVersionService = &stateVersions{client: client}
	client.WorkspaceService = &workspaces{client: client}
	return client, nil
}

type rawAPIMetadata struct {
	// APIVersion is the raw API version string reported by the server in the
	// TFP-API-Version response header, or an empty string if that header
	// field was not included in the response.
	APIVersion string
	// RateLimit is the raw API version string reported by the server in the
	// X-RateLimit-Limit response header, or an empty string if that header
	// field was not included in the response.
	RateLimit string
}
