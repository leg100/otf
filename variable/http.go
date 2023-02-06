package variable

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app service
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
//
func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/vars", h.create).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", h.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.delete).Methods("DELETE")
}

// VariableList assembles a workspace list JSONAPI DTO
type VariableList struct {
	variables []*Variable
}

func (l *VariableList) ToJSONAPI() any {
	variables := &jsonapiList{}
	for _, v := range l.variables {
		variables.Items = append(variables.Items, v.ToJSONAPI().(*jsonapiVariable))
	}
	return variables
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts jsonapiVariableCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := h.app.create(r.Context(), workspaceID, otf.CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*otf.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, variable, jsonapi.WithCode(http.StatusCreated))
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := h.app.get(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, variable)
}

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variables, err := h.app.list(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &VariableList{variables})
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts jsonapiVariableUpdateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	updated, err := h.app.update(r.Context(), variableID, otf.UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*otf.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, updated)
}

func (h *handlers) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = h.app.delete(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
}
