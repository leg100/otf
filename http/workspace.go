package http

import (
	"net/http"
	"strings"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

// Workspace assembles a workspace JSONAPI DTO
type Workspace struct {
	r *http.Request
	otf.Application
	*otf.Workspace
}

// workspaceNameParams are those parameters used when looking up a workspace by
// name
type workspaceNameParams struct {
	Name         string `schema:"workspace_name,required"`
	Organization string `schema:"organization_name,required"`
}

func (ws *Workspace) ToJSONAPI() any {
	subject, err := otf.SubjectFromContext(ws.r.Context())
	if err != nil {
		panic(err.Error())
	}
	perms, err := ws.ListWorkspacePermissions(ws.r.Context(), ws.ID())
	if err != nil {
		panic(err.Error())
	}
	policy := &otf.WorkspacePolicy{
		Organization: ws.Organization(),
		WorkspaceID:  ws.ID(),
		Permissions:  perms,
	}

	obj := &dto.Workspace{
		ID: ws.ID(),
		Actions: &dto.WorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     ws.AllowDestroyPlan(),
		AutoApply:            ws.AutoApply(),
		CanQueueDestroyPlan:  ws.CanQueueDestroyPlan(),
		CreatedAt:            ws.CreatedAt(),
		Description:          ws.Description(),
		Environment:          ws.Environment(),
		ExecutionMode:        string(ws.ExecutionMode()),
		FileTriggersEnabled:  ws.FileTriggersEnabled(),
		GlobalRemoteState:    ws.GlobalRemoteState(),
		Locked:               ws.Locked(),
		MigrationEnvironment: ws.MigrationEnvironment(),
		Name:                 ws.Name(),
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations: ws.ExecutionMode() == "remote",
		Permissions: &dto.WorkspacePermissions{
			CanLock:           subject.CanAccessWorkspace(otf.LockWorkspaceAction, policy),
			CanUnlock:         subject.CanAccessWorkspace(otf.UnlockWorkspaceAction, policy),
			CanForceUnlock:    subject.CanAccessWorkspace(otf.UnlockWorkspaceAction, policy),
			CanQueueApply:     subject.CanAccessWorkspace(otf.ApplyRunAction, policy),
			CanQueueDestroy:   subject.CanAccessWorkspace(otf.ApplyRunAction, policy),
			CanQueueRun:       subject.CanAccessWorkspace(otf.CreateRunAction, policy),
			CanDestroy:        subject.CanAccessWorkspace(otf.DeleteWorkspaceAction, policy),
			CanReadSettings:   subject.CanAccessWorkspace(otf.GetWorkspaceAction, policy),
			CanUpdate:         subject.CanAccessWorkspace(otf.UpdateWorkspaceAction, policy),
			CanUpdateVariable: subject.CanAccessWorkspace(otf.UpdateWorkspaceAction, policy),
		},
		QueueAllRuns:               ws.QueueAllRuns(),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		SourceName:                 ws.SourceName(),
		SourceURL:                  ws.SourceURL(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           ws.TerraformVersion(),
		TriggerPrefixes:            ws.TriggerPrefixes(),
		WorkingDirectory:           ws.WorkingDirectory(),
		UpdatedAt:                  ws.UpdatedAt(),
		Organization:               &dto.Organization{Name: ws.Organization()},
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: limit support to organization, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := ws.r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				org, err := ws.GetOrganization(ws.r.Context(), ws.Organization())
				if err != nil {
					panic(err.Error()) // throws HTTP500
				}
				obj.Organization = (&Organization{org}).ToJSONAPI().(*dto.Organization)
			}
		}
	}
	return obj
}

// WorkspaceList assembles a workspace list JSONAPI DTO
type WorkspaceList struct {
	r *http.Request
	otf.Application
	*otf.WorkspaceList
}

func (l *WorkspaceList) ToJSONAPI() any {
	obj := &dto.WorkspaceList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Workspace{l.r, l.Application, item}).ToJSONAPI().(*dto.Workspace))
	}
	return obj
}

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts dto.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.CreateWorkspace(r.Context(), otf.CreateWorkspaceOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*otf.ExecutionMode)(opts.ExecutionMode),
		FileTriggersEnabled:        opts.FileTriggersEnabled,
		GlobalRemoteState:          opts.GlobalRemoteState,
		MigrationEnvironment:       opts.MigrationEnvironment,
		Name:                       opts.Name,
		Organization:               opts.Organization,
		QueueAllRuns:               opts.QueueAllRuns,
		SpeculativeEnabled:         opts.SpeculativeEnabled,
		SourceName:                 opts.SourceName,
		SourceURL:                  opts.SourceURL,
		StructuredRunOutputEnabled: opts.StructuredRunOutputEnabled,
		TerraformVersion:           opts.TerraformVersion,
		TriggerPrefixes:            opts.TriggerPrefixes,
		WorkingDirectory:           opts.WorkingDirectory,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws}, withCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspace(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) GetWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
	if err := decode.All(&params, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	wsl, err := s.Application.ListWorkspaces(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &WorkspaceList{r, s.Application, wsl})
}

// UpdateWorkspace updates a workspace using its ID.
//
// TODO: support updating workspace's vcs repo.
func (s *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	s.updateWorkspace(w, r, workspaceID)
}

// UpdateWorkspaceByName updates a workspace using its name and organization.
//
// TODO: support updating workspace's vcs repo.
func (s *Server) UpdateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
	if err := decode.Route(&params, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	s.updateWorkspace(w, r, ws.ID())
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := otf.WorkspaceLockOptions{}
	ws, err := s.Application.LockWorkspace(r.Context(), id, opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := otf.WorkspaceUnlockOptions{}
	ws, err := s.Application.UnlockWorkspace(r.Context(), id, opts)
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = s.Application.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
	if err := decode.All(&params, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	_, err = s.Application.DeleteWorkspace(r.Context(), ws.ID())
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	opts := dto.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.UpdateWorkspace(r.Context(), workspaceID, otf.UpdateWorkspaceOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*otf.ExecutionMode)(opts.ExecutionMode),
		FileTriggersEnabled:        opts.FileTriggersEnabled,
		GlobalRemoteState:          opts.GlobalRemoteState,
		Name:                       opts.Name,
		QueueAllRuns:               opts.QueueAllRuns,
		SpeculativeEnabled:         opts.SpeculativeEnabled,
		StructuredRunOutputEnabled: opts.StructuredRunOutputEnabled,
		TerraformVersion:           opts.TerraformVersion,
		TriggerPrefixes:            opts.TriggerPrefixes,
		WorkingDirectory:           opts.WorkingDirectory,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{r, s.Application, ws})
}
