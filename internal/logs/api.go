package logs

import (
	"bytes"
	"io"
	"net/http"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
)

type api struct {
	internal.Verifier // for verifying upload url

	svc *Service
}

func (a *api) addHandlers(r *mux.Router) {
	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Verifier))
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", a.getLogs).Methods("GET")

	// client is typically otf-agent
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()
	r.HandleFunc("/runs/{run_id}/logs/{phase}", a.putLogs).Methods("PUT")
}

func (a *api) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	chunk, err := a.svc.GetChunk(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if _, err := w.Write(chunk.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *api) putLogs(w http.ResponseWriter, r *http.Request) {
	var opts PutChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts.Data = buf.Bytes()
	if err := a.svc.PutChunk(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
