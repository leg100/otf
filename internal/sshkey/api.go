package sshkey

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	api struct {
		*Service
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()

	r.HandleFunc("/ssh-keys/{ssh_key_id}", a.getPrivateKey).Methods("GET")
}

func (a *api) getPrivateKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("ssh_key_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := a.GetPrivateKey(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.Write(key)
}
