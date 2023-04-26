package workspace

import (
	"context"
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

type (
	// jsonapiMarshaler marshals workspace into a struct suitable for marshaling
	// into json-api
	jsonapiMarshaler struct {
		OrganizationService
		PermissionsService
	}

	outputsJSONAPIService interface {
		GetCurrentOutputsJSONAPI(ctx context.Context, workspaceID string) (*jsonapi.StateVersionOutputList, error)
	}
)

func (m *jsonapiMarshaler) toWorkspace(ws *Workspace, r *http.Request) (*jsonapi.Workspace, error) {
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

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: support is currently limited to a couple of included resources...
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				org, err = m.GetOrganizationJSONAPI(r.Context(), ws.Organization)
				if err != nil {
					return nil, err
				}
			case "outputs":
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
	}, nil
}

func (m *jsonapiMarshaler) toList(list *WorkspaceList, r *http.Request) (*jsonapi.WorkspaceList, error) {
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
