package logs

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

type api struct {
	internal.Verifier // for verifying upload url

	svc Service
}

func (a *api) addHandlers(r *mux.Router) {
	// client is typically terraform-cli
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Verifier))
	signed.HandleFunc("/runs/{run_id}/logs/{phase}", a.getLogs).Methods("GET")

	r = otfhttp.APIRouter(r)

	// client is typically an external agent
	r.HandleFunc("/runs/{run_id}/logs/{phase}", a.putLogs).Methods("PUT")
}

func (a *api) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts internal.GetChunkOptions
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
	var opts internal.PutChunkOptions
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
