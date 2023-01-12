package http

import (
	"net/http"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

// Variable assembles a workspace JSONAPI DTO
type Variable struct {
	*otf.Variable
}

func (v *Variable) ToJSONAPI() any {
	to := dto.Variable{
		ID:          v.ID(),
		Key:         v.Key(),
		Value:       v.Value(),
		Description: v.Description(),
		Category:    string(v.Category()),
		Sensitive:   v.Sensitive(),
		HCL:         v.HCL(),
		Workspace: &dto.Workspace{
			ID: v.WorkspaceID(),
		},
	}
	if to.Sensitive {
		to.Value = "" // scrub sensitive values
	}
	return &to
}

// VariableList assembles a workspace list JSONAPI DTO
type VariableList struct {
	variables []*otf.Variable
}

func (l *VariableList) ToJSONAPI() any {
	variables := &dto.VariableList{}
	for _, v := range l.variables {
		variables.Items = append(variables.Items, (&Variable{v}).ToJSONAPI().(*dto.Variable))
	}
	return variables
}

func (s *Server) CreateVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts dto.VariableCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := s.Application.CreateVariable(r.Context(), workspaceID, otf.CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*otf.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Variable{variable}, withCode(http.StatusCreated))
}

func (s *Server) GetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	variable, err := s.Application.GetVariable(r.Context(), variableID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Variable{variable})
}

func (s *Server) ListVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	variables, err := s.Application.ListVariables(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &VariableList{variables})
}

func (s *Server) UpdateVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts dto.VariableUpdateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	updated, err := s.Application.UpdateVariable(r.Context(), variableID, otf.UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*otf.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Variable{updated})
}

func (s *Server) DeleteVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = s.Application.DeleteVariable(r.Context(), variableID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}
