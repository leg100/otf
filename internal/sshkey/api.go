package sshkey

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	Client apiClient
}

type apiClient interface {
	GetSSHKeyPrivateKey(ctx context.Context, id resource.ID) ([]byte, error)
}

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/ssh-keys/{ssh_key_id}", a.getPrivateKey).Methods("GET")
}

func (a *API) getPrivateKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("ssh_key_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := a.Client.GetSSHKeyPrivateKey(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.Write(key)
}
