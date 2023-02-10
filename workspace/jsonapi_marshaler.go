package workspace

import (
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// JSONAPIMarshaler marshals workspace into a struct suitable for marshaling
// into json-api
type JSONAPIMarshaler struct {
	otf.Application
}

func (m *JSONAPIMarshaler) toJSONAPI(ws *Workspace, r *http.Request) (*jsonapi.Workspace, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	perms, err := m.ListWorkspacePermissions(r.Context(), ws.id)
	if err != nil {
		return nil, err
	}
	policy := &otf.WorkspacePolicy{
		Organization: ws.Organization(),
		WorkspaceID:  ws.ID(),
		Permissions:  perms,
	}

	org := &jsonapi.Organization{Name: ws.Organization()}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: limit support to organization, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				org, err = m.GetOrganizationJSONAPI(r.Context(), ws.Organization())
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &jsonapi.Workspace{
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
		Organization:               org,
	}, nil
}

func (m *JSONAPIMarshaler) toJSONAPIList(list *WorkspaceList, r *http.Request) (*jsonapi.WorkspaceList, error) {
	var items []*jsonapi.Workspace
	for _, ws := range list.Items {
		item, err := m.toJSONAPI(ws, r)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &jsonapi.WorkspaceList{
		Pagination: list.Pagination.ToJSONAPI(),
	}, nil
}
