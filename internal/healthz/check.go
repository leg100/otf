package healthz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	ok    status = "OK"
	notok status = "UNHEALTHY"
)

type status string

type Check struct {
	Client CheckClient
}

type CheckClient interface {
	Ping(ctx context.Context) error
}

func (c *Check) AddHandlers(r *mux.Router) {
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		var response struct {
			Status status
			Error  string `json:"omitempty"`
		}

		w.Header().Set("Content-type", "application/json")

		if err := c.Client.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)

			response.Status = notok
			response.Error = err.Error()
		} else {
			response.Status = ok
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			w.Write(fmt.Appendf(nil, "unable to encode health check response to json: %s", err.Error()))
		}
	})
}
