package sshkey

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type TFEAPI struct {
	*tfeapi.Responder
	Client tfeClient
}

type tfeClient interface {
	CreateSSHKey(ctx context.Context, opts CreateOptions) (*SSHKey, error)
	UpdateSSHKey(ctx context.Context, id resource.TfeID, opts UpdateOptions) (*SSHKey, error)
	ListSSHKeys(ctx context.Context, org organization.Name) ([]*SSHKey, error)
	GetSSHKey(ctx context.Context, id resource.TfeID) (*SSHKey, error)
	DeleteSSHKey(ctx context.Context, id resource.TfeID) (*SSHKey, error)
}

// TFESSHKey represents an SSH key in the TFE API.
type TFESSHKey struct {
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

func (a *TFEAPI) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/ssh-keys", a.createSSHKey).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/ssh-keys", a.listSSHKeys).Methods("GET")
	r.HandleFunc("/ssh-keys/{id}", a.getSSHKey).Methods("GET")
	r.HandleFunc("/ssh-keys/{id}", a.updateSSHKey).Methods("PATCH")
	r.HandleFunc("/ssh-keys/{id}", a.deleteSSHKey).Methods("DELETE")
}

func (a *TFEAPI) createSSHKey(w http.ResponseWriter, r *http.Request) {
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
	key, err := a.Client.CreateSSHKey(r.Context(), CreateOptions{
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

func (a *TFEAPI) listSSHKeys(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.Route(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	keys, err := a.Client.ListSSHKeys(r.Context(), pathParams.Organization)
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

func (a *TFEAPI) getSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	key, err := a.Client.GetSSHKey(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(key), http.StatusOK)
}

func (a *TFEAPI) updateSSHKey(w http.ResponseWriter, r *http.Request) {
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
	key, err := a.Client.UpdateSSHKey(r.Context(), id, UpdateOptions{
		Name: params.Name,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(key), http.StatusOK)
}

func (a *TFEAPI) deleteSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.Client.DeleteSSHKey(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convert(from *SSHKey) *TFESSHKey {
	return &TFESSHKey{
		ID:   from.ID,
		Name: from.Name,
	}
}
