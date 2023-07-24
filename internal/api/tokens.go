package api

import (
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tokens"
)

func (a *api) addTokenHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Agent token routes
	r.HandleFunc("/agent/details", a.getCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", a.createAgentToken).Methods("POST")

	// Run token routes
	r.HandleFunc("/tokens/run/create", a.createRunToken).Methods("POST")

	// Organization token routes
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.createOrganizationToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.getOrganizationToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.deleteOrganizationToken).Methods("DELETE")
}

func (a *api) createRunToken(w http.ResponseWriter, r *http.Request) {
	var opts types.CreateRunTokenOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	token, err := a.CreateRunToken(r.Context(), tokens.CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		Error(w, err)
		return
	}

	w.Write(token)
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts types.AgentTokenCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	token, err := a.CreateAgentToken(r.Context(), tokens.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := tokens.AgentFromContext(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	b, err := jsonapi.Marshal(&types.AgentToken{
		ID:           at.ID,
		Organization: at.Organization,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.Write(b)
}

func (a *api) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var opts types.OrganizationTokenCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	ot, token, err := a.CreateOrganizationToken(r.Context(), tokens.CreateOrganizationTokenOptions{
		Organization: org,
		Expiry:       opts.ExpiredAt,
	})
	if err != nil {
		Error(w, err)
		return
	}

	b, err := jsonapi.Marshal(&types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		Token:     string(token),
		ExpiredAt: ot.Expiry,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.Write(b)
}

func (a *api) getOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	ot, err := a.GetOrganizationToken(r.Context(), org)
	if err != nil {
		Error(w, err)
		return
	}
	if ot == nil {
		Error(w, internal.ErrResourceNotFound)
		return
	}

	b, err := jsonapi.Marshal(&types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		ExpiredAt: ot.Expiry,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.Write(b)
}

func (a *api) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	err = a.DeleteOrganizationToken(r.Context(), org)
	if err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
