// Package tfeapi provides common functionality useful for implementation of the
// Hashicorp TFE/TFC API, which uses the json:api encoding
package tfeapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
)

const (
	// APIPrefixV2 is the URL path prefix for TFE API endpoints
	APIPrefixV2 = "/api/v2/"
	// ModuleV1Prefix is the URL path prefix for module registry endpoints
	ModuleV1Prefix = "/v1/modules/"
)

// errUnmarshal wraps errors resulting from a failure to unmarshal request
// parameters.
var errUnmarshal = errors.New("error unmarshalling request parameters")

func Unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := jsonapi.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%w: %w", errUnmarshal, err)
	}
	return nil
}

type Handlers struct {
	Handlers []internal.Handlers
	Verifier Verifier
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	apiRouter := r.PathPrefix(APIPrefixV2).Subrouter()

	// Middleware to alter requests/responses
	apiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add TFP API version header to every API response.
			//
			// Version 2.5 is the minimum version terraform requires for the
			// newer 'cloud' configuration block:
			// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
			w.Header().Set("TFP-API-Version", "2.5")

			// Remove trailing slash from all requests
			//
			// TODO: does this actually do anything? The mux router will return
			// a 404 for any paths not matching a handler *before* the request
			// is sent to this middleware, so if all handlers don't have a
			// trailing slash then a 404 will be returned.
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

			next.ServeHTTP(w, r)
		})
	})

	// Handle ping sent by go-tfe upon initialization
	apiRouter.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Verify signed URLs
	signedRouter := r.PathPrefix(SignedPrefixWithSignature).Subrouter()
	signedRouter.Use(verifySignedURL(h.Verifier))

	// Register each handler, both with an API prefix and with a signed URL
	// prefix (to allow each route to be accessible with a signed URL as well).
	for _, h := range h.Handlers {
		h.AddHandlers(apiRouter)
		h.AddHandlers(signedRouter)
	}
}
