package sshkey

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/tfeapi"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/organizations/{organization_name}/ssh-keys", a.createSSHKey).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/ssh-keys", a.listSSHKeys).Methods("GET")
	r.HandleFunc("/ssh-keys/{id}", a.getSSHKey).Methods("GET")
	r.HandleFunc("/ssh-keys/{id}", a.updateSSHKey).Methods("PATCH")
	r.HandleFunc("/ssh-keys/{id}", a.deleteSSHKey).Methods("DELETE")
}

func (a *tfe) createSSHKey(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.Route(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFESSHKeyCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Create(r.Context(), CreateOptions{
		Organization: &pathParams.Organization,
		Name:         params.Name,
		PrivateKey:   params.Value,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(key), http.StatusCreated)
}

func (a *tfe) listSSHKeys(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.Route(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	keys, err := a.List(r.Context(), pathParams.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	to := make([]*TFESSHKey, len(keys))
	for i, k := range keys {
		to[i] = a.convert(k)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Get(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(key), http.StatusOK)
}

func (a *tfe) updateSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFESSHKeyUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Update(r.Context(), id, UpdateOptions{
		Name:       params.Name,
		PrivateKey: params.Value,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(key), http.StatusOK)
}

func (a *tfe) deleteSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Delete(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *SSHKey) *TFESSHKey {
	return &TFESSHKey{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
		Name:      from.Name,
		Value:     from.PrivateKey,
	}
}
