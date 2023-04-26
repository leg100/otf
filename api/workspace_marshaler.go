package api

import (
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/workspace"
)

func (m *jsonapiMarshaler) toWorkspace(ws *workspace.Workspace, r *http.Request) (*jsonapi.Workspace, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	policy, err := m.GetPolicy(r.Context(), ws.ID)
	if err != nil {
		return nil, err
	}
	perms := &jsonapi.WorkspacePermissions{
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
	}

	org := &jsonapi.Organization{Name: ws.Organization}
	outputs := []*jsonapi.StateVersionOutput{}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: support is currently limited to a couple of included resources...
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				unmarshaled, err := m.GetOrganization(r.Context(), ws.Organization)
				if err != nil {
					return nil, err
				}
				org = m.toOrganization(unmarshaled)
			case "outputs":
				sv, err := m.GetCurrentStateVersion(r.Context(), ws.ID)
				if err != nil {
					return nil, err
				}
				for _, out := range sv.Outputs {
					outputs = append(outputs, m.toOutput(out))
				}
			}
		}
	}
	return &jsonapi.Workspace{
		ID: ws.ID,
		Actions: &jsonapi.WorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     ws.AllowDestroyPlan,
		AutoApply:            ws.AutoApply,
		CanQueueDestroyPlan:  ws.CanQueueDestroyPlan,
		CreatedAt:            ws.CreatedAt,
		Description:          ws.Description,
		Environment:          ws.Environment,
		ExecutionMode:        string(ws.ExecutionMode),
		FileTriggersEnabled:  ws.FileTriggersEnabled,
		GlobalRemoteState:    ws.GlobalRemoteState,
		Locked:               ws.Locked(),
		MigrationEnvironment: ws.MigrationEnvironment,
		Name:                 ws.Name,
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations:                 ws.ExecutionMode == "remote",
		Permissions:                perms,
		QueueAllRuns:               ws.QueueAllRuns,
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		SourceName:                 ws.SourceName,
		SourceURL:                  ws.SourceURL,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           ws.TerraformVersion,
		TriggerPrefixes:            ws.TriggerPrefixes,
		WorkingDirectory:           ws.WorkingDirectory,
		UpdatedAt:                  ws.UpdatedAt,
		Organization:               org,
		Outputs:                    outputs,
	}, nil
}

func (m *jsonapiMarshaler) toWorkspaceList(list *workspace.WorkspaceList, r *http.Request) (*jsonapi.WorkspaceList, error) {
	var items []*jsonapi.Workspace
	for _, ws := range list.Items {
		item, err := m.toWorkspace(ws, r)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &jsonapi.WorkspaceList{
		Items:      items,
		Pagination: jsonapi.NewPagination(list.Pagination),
	}, nil
}
