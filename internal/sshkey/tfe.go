package sshkey

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

// tfeSSHKey represents an SSH key in the TFE API.
type tfeSSHKey struct {
	ID        resource.TfeID `jsonapi:"primary,ssh-keys"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created-at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated-at"`
	Name      string         `jsonapi:"attribute" json:"name"`
}

// tfeSSHKeyCreateOptions are the options for creating a new SSH key via the TFE API.
type tfeSSHKeyCreateOptions struct {
	// Type is used by JSON:API to set the resource type.
	Type string `jsonapi:"primary,ssh-keys"`

	// The name of the SSH key.
	Name string `jsonapi:"attribute" json:"name"`

	// The private key value (PEM-encoded). Write-only.
	Value string `jsonapi:"attribute" json:"value"`
}

// tfeSSHKeyUpdateOptions are the options for updating an SSH key via the TFE API.
type tfeSSHKeyUpdateOptions struct {
	// Type is used by JSON:API to set the resource type.
	Type string `jsonapi:"primary,ssh-keys"`

	// The name of the SSH key.
	Name *string `jsonapi:"attribute" json:"name,omitempty"`
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
	var params tfeSSHKeyCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Create(r.Context(), CreateOptions{
		Organization: pathParams.Organization,
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
	to := make([]*tfeSSHKey, len(keys))
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
	var params tfeSSHKeyUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Update(r.Context(), id, UpdateOptions{
		Name: params.Name,
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
	_, err = a.Delete(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *SSHKey) *tfeSSHKey {
	return &tfeSSHKey{
		ID:   from.ID,
		Name: from.Name,
	}
}
