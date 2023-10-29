// Package tfeapi provides common functionality useful for implementation of the
// Hashicorp TFE/TFC API, which uses the json:api encoding
package tfeapi

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
)

const (
	// APIPrefixV2 is the URL path prefix for TFE API endpoints
	APIPrefixV2 = "/api/v2/"
	// ModuleV1Prefix is the URL path prefix for module registry endpoints
	ModuleV1Prefix = "/v1/modules/"
)

func Unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return jsonapi.Unmarshal(b, v)
}

type Handlers struct{}

func (h *Handlers) AddHandlers(r *mux.Router) {
	// handle ping sent by go-tfe upon initialization
	r.HandleFunc(path.Join(APIPrefixV2, "ping"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Middleware to alter requests/responses
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// API requests/responses
			if strings.HasPrefix(r.URL.Path, APIPrefixV2) {
				// Add TFP API version header to every API response.
				//
				// Version 2.5 is the minimum version terraform requires for the
				// newer 'cloud' configuration block:
				// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
				w.Header().Set("TFP-API-Version", "2.5")
			}

			// Remove trailing slash from all requests
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

			next.ServeHTTP(w, r)
		})
	})

}
