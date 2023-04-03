package orgcreator

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
)

type api struct {
	svc Service

	*organization.JSONAPIMarshaler
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations", h.CreateOrganization).Methods("POST")
}

func (h *api) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.OrganizationCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}

	org, err := h.svc.CreateOrganization(r.Context(), OrganizationCreateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	jsonapi.WriteResponse(w, r, h.ToOrganization(org), jsonapi.WithCode(http.StatusCreated))
}
