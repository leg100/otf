package client

import "net/http"

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

// RetryLogHook allows a function to run before each retry.
type RetryLogHook func(attemptNum int, resp *http.Response)
