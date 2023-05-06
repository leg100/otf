package internal

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
)

// JSONAPIClient is a client capable of interacting with a json-api API
type JSONAPIClient interface {
	// NewRequest constructs a new json-api request
	NewRequest(method, path string, params any) (*retryablehttp.Request, error)
	// Do sends a json-api request and populates v with a json-api response.
	Do(ctx context.Context, req *retryablehttp.Request, v any) error
}
