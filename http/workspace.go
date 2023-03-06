package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
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

	obj := &jsonapi.Workspace{
		ID: ws.ID(),
		Actions: &jsonapi.WorkspaceActions{
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
		Permissions: &jsonapi.WorkspacePermissions{
			CanLock:           subject.CanAccessWorkspace(rbac.LockWorkspaceAction, policy),
			CanUnlock:         subject.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy),
			CanForceUnlock:    subject.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy),
			CanQueueApply:     subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
			CanQueueDestroy:   subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
			CanQueueRun:       subject.CanAccessWorkspace(rbac.CreateRunAction, policy),
			CanDestroy:        subject.CanAccessWorkspace(rbac.DeleteWorkspaceAction, policy),
			CanReadSettings:   subject.CanAccessWorkspace(rbac.GetWorkspaceAction, policy),
			CanUpdate:         subject.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
			CanUpdateVariable: subject.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
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
		Organization:               &jsonapi.Organization{Name: ws.Organization()},
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
				obj.Organization = (&Organization{org}).ToJSONAPI().(*jsonapi.Organization)
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
	obj := &jsonapi.WorkspaceList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Workspace{l.r, l.Application, item}).ToJSONAPI().(*jsonapi.Workspace))
	}
	return obj
}

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var params jsonapi.WorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	opts := otf.CreateWorkspaceOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*otf.ExecutionMode)(params.ExecutionMode),
		FileTriggersEnabled:        params.FileTriggersEnabled,
		GlobalRemoteState:          params.GlobalRemoteState,
		MigrationEnvironment:       params.MigrationEnvironment,
		Name:                       params.Name,
		Organization:               params.Organization,
		QueueAllRuns:               params.QueueAllRuns,
		SpeculativeEnabled:         params.SpeculativeEnabled,
		SourceName:                 params.SourceName,
		SourceURL:                  params.SourceURL,
		StructuredRunOutputEnabled: params.StructuredRunOutputEnabled,
		TerraformVersion:           params.TerraformVersion,
		TriggerPrefixes:            params.TriggerPrefixes,
		WorkingDirectory:           params.WorkingDirectory,
	}
	if params.Operations != nil {
		if params.ExecutionMode != nil {
			err := errors.New("operations is deprecated and cannot be specified when execution mode is used")
			writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		if *params.Operations {
			opts.ExecutionMode = otf.ExecutionModePtr(otf.RemoteExecutionMode)
		} else {
			opts.ExecutionMode = otf.ExecutionModePtr(otf.LocalExecutionMode)
		}
	}

	if params.VCSRepo != nil {
		if params.VCSRepo.Identifier == nil || params.VCSRepo.OAuthTokenID == nil {
			err := errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo")
			writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		opts.Repo = &otf.Connection{
			Identifier:    *params.VCSRepo.Identifier,
			VCSProviderID: *params.VCSRepo.OAuthTokenID,
		}
		if params.VCSRepo.Branch != nil {
			opts.Branch = params.VCSRepo.Branch
		}
	}
	ws, err := otf.CreateWorkspace(r.Context(), s.Application, opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws}, withCode(http.StatusCreated))
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
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
	jsonapi.WriteResponse(w, r, &WorkspaceList{r, s.Application, wsl})
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
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
	opts := jsonapi.WorkspaceUpdateOptions{}
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}
