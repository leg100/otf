package api

import (
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/workspace"
)

func (m *jsonapiMarshaler) toWorkspace(from *workspace.Workspace, r *http.Request) (*jsonapi.Workspace, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	policy, err := m.GetPolicy(r.Context(), from.ID)
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

	to := &jsonapi.Workspace{
		ID: from.ID,
		Actions: &jsonapi.WorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     from.AllowDestroyPlan,
		AutoApply:            from.AutoApply,
		CanQueueDestroyPlan:  from.CanQueueDestroyPlan,
		CreatedAt:            from.CreatedAt,
		Description:          from.Description,
		Environment:          from.Environment,
		ExecutionMode:        string(from.ExecutionMode),
		FileTriggersEnabled:  from.FileTriggersEnabled,
		GlobalRemoteState:    from.GlobalRemoteState,
		Locked:               from.Locked(),
		MigrationEnvironment: from.MigrationEnvironment,
		Name:                 from.Name,
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations:                 from.ExecutionMode == "remote",
		Permissions:                perms,
		QueueAllRuns:               from.QueueAllRuns,
		SpeculativeEnabled:         from.SpeculativeEnabled,
		SourceName:                 from.SourceName,
		SourceURL:                  from.SourceURL,
		StructuredRunOutputEnabled: from.StructuredRunOutputEnabled,
		TerraformVersion:           from.TerraformVersion,
		TriggerPrefixes:            from.TriggerPrefixes,
		WorkingDirectory:           from.WorkingDirectory,
		UpdatedAt:                  from.UpdatedAt,
		Organization:               &jsonapi.Organization{Name: from.Organization},
		Outputs:                    []*jsonapi.StateVersionOutput{},
	}

	if from.LatestRun != nil {
		to.CurrentRun = &jsonapi.Run{ID: from.LatestRun.ID}
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: support is currently limited to a couple of included resources...
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				unmarshaled, err := m.GetOrganization(r.Context(), from.Organization)
				if err != nil {
					return nil, err
				}
				to.Organization = m.toOrganization(unmarshaled)
			case "outputs":
				sv, err := m.GetCurrentStateVersion(r.Context(), from.ID)
				if err != nil {
					return nil, err
				}
				for _, out := range sv.Outputs {
					to.Outputs = append(to.Outputs, m.toOutput(out))
				}
			}
		}
	}

	return to, nil
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
