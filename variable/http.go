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
	r.HandleFunc("/workspaces/{workspace_id}/vars", h.CreateVariable)
	r.HandleFunc("/workspaces/{workspace_id}/vars", h.ListVariables)
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.GetVariable)
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.UpdateVariable)
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", h.DeleteVariable)
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

func (h *handlers) CreateVariable(w http.ResponseWriter, r *http.Request) {
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

func (s *handlers) GetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := s.app.get(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, variable)
}

func (s *handlers) ListVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	variables, err := s.app.list(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &VariableList{variables})
}

func (s *handlers) UpdateVariable(w http.ResponseWriter, r *http.Request) {
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
	updated, err := s.app.update(r.Context(), variableID, otf.UpdateVariableOptions{
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

func (s *handlers) DeleteVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = s.app.delete(r.Context(), variableID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
}
