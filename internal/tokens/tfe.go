package tokens

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type tfe struct {
	TokensService
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
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

func (a *tfe) createRunToken(w http.ResponseWriter, r *http.Request) {
	var opts types.CreateRunTokenOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	token, err := a.CreateRunToken(r.Context(), CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.Write(token)
}

func (a *tfe) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts types.AgentTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.CreateAgentToken(r.Context(), CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *tfe) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := AgentFromContext(r.Context())
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	to := &types.AgentToken{
		ID:           at.ID,
		Organization: at.Organization,
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.OrganizationTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, token, err := a.CreateOrganizationToken(r.Context(), CreateOrganizationTokenOptions{
		Organization: org,
		Expiry:       opts.ExpiredAt,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := &types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		Token:     string(token),
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) getOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, err := a.GetOrganizationToken(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if ot == nil {
		tfeapi.Error(w, internal.ErrResourceNotFound)
		return
	}

	to := &types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	err = a.DeleteOrganizationToken(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
