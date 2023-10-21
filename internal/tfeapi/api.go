// Package tfeapi provides common functionality useful for implementation of the
// Hashicorp TFE/TFC API, which uses the json:api encoding
package tfeapi

import (
	"io"
	"net/http"
	"path"

	otfhttp "github.com/leg100/otf/internal/http"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
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
	// terraform go-tfe client initialization sends a ping
	r.HandleFunc(path.Join(otfhttp.APIPrefixV2, "ping"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
