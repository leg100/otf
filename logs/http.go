package logs

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
)

type handlers struct {
	app
	otf.Verifier // for verifying upload url
}

func newHandlers(opts handlersOptions) *handlers {
	return &handlers{
		app:      opts.app,
		Verifier: opts.Verifier,
	}
}

type handlersOptions struct {
	app
	max int64
	otf.Verifier
}

func (s *handlers) AddHandlers(r *mux.Router) {
	// client is typically an external agent
	r.HandleFunc("/runs/{run_id}/logs/{phase}", s.putLogs).Methods("PUT")

	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use((&otfhttp.SignatureVerifier{s.Verifier}).Handler)
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", s.getLogs).Methods("GET")
}

func (s *handlers) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	chunk, err := s.GetChunk(r.Context(), opts)
	// ignore not found errors because terraform-cli may call this endpoint
	// before any logs have been written and it'll exit with an error if a not
	// found error is received.
	if err != nil && err != otf.ErrResourceNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if _, err := w.Write(chunk.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *handlers) putLogs(w http.ResponseWriter, r *http.Request) {
	chunk := Chunk{}
	if err := decode.All(&chunk, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	chunk.Data = buf.Bytes()
	if err := s.PutChunk(r.Context(), chunk); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
