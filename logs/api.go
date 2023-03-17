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

type api struct {
	otf.Verifier // for verifying upload url

	svc Service
}

func (a *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// client is typically an external agent
	r.HandleFunc("/runs/{run_id}/logs/{phase}", a.putLogs).Methods("PUT")

	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(otf.VerifySignedURL(a.Verifier))
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", a.getLogs).Methods("GET")
}

func (a *api) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts otf.GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	chunk, err := a.svc.GetChunk(r.Context(), opts)
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
	if err := s.svc.PutChunk(r.Context(), chunk); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
