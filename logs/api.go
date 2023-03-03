package logs

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

type api struct {
	service
	otf.Verifier // for verifying upload url
}

func (s *api) addHandlers(r *mux.Router) {
	// client is typically an external agent
	r.HandleFunc("/runs/{run_id}/logs/{phase}", s.putLogs).Methods("PUT")

	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(otf.VerifySignedURL(s.Verifier))
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", s.getLogs).Methods("GET")
}

func (s *api) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts otf.GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	chunk, err := s.service.GetChunk(r.Context(), opts)
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

func (s *api) putLogs(w http.ResponseWriter, r *http.Request) {
	chunk := otf.Chunk{}
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
	if err := s.service.PutChunk(r.Context(), chunk); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
