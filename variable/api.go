package variable

import (
	"fmt"
	"net/http"

	otfhttp "github.com/leg100/otf/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	svc             Service
	tokenMiddleware mux.MiddlewareFunc
}

// Implements TFC workspace variables API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables#update-variables
func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)
	r.Use(h.tokenMiddleware) // require bearer token

	r.HandleFunc("/workspaces/{workspace_id}/vars", h.create).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", h.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.delete).Methods("DELETE")
}

func (h *api) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts jsonapi.VariableCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := h.svc.create(r.Context(), workspaceID, CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	h.writeResponse(w, r, variable, jsonapi.WithCode(http.StatusCreated))
}

func (h *api) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := h.svc.get(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	h.writeResponse(w, r, variable)
}

func (h *api) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variables, err := h.svc.ListVariables(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	h.writeResponse(w, r, variables)
}

func (h *api) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts jsonapi.VariableUpdateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	updated, err := h.svc.update(r.Context(), variableID, UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	h.writeResponse(w, r, updated)
}

func (h *api) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = h.svc.delete(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
}

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (h *api) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var payload any

	convert := func(from *Variable) *jsonapi.Variable {
		to := jsonapi.Variable{
			ID:          from.ID,
			Key:         from.Key,
			Value:       from.Value,
			Description: from.Description,
			Category:    string(from.Category),
			Sensitive:   from.Sensitive,
			HCL:         from.HCL,
			Workspace: &jsonapi.Workspace{
				ID: from.WorkspaceID,
			},
		}
		if to.Sensitive {
			to.Value = "" // scrub sensitive values
		}
		return &to
	}

	switch v := v.(type) {
	case *Variable:
		payload = convert(v)
	case []*Variable:
		var to jsonapi.VariableList
		for _, from := range v {
			to.Items = append(to.Items, convert(from))
		}
		payload = &to
	default:
		err := fmt.Errorf("no json:api struct found for %T", v)
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, payload, opts...)
}
