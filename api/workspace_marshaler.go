package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/workspace"
)

func (m *jsonapiMarshaler) toWorkspace(from *workspace.Workspace, r *http.Request) (*types.Workspace, []jsonapi.MarshalOption, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, nil, err
	}
	policy, err := m.GetPolicy(r.Context(), from.ID)
	if err != nil {
		return nil, nil, err
	}
	perms := &types.WorkspacePermissions{
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

	to := &types.Workspace{
		ID: from.ID,
		Actions: &types.WorkspaceActions{
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
		TagNames:                   from.Tags,
		UpdatedAt:                  from.UpdatedAt,
		Organization:               &types.Organization{Name: from.Organization},
		Outputs:                    []*types.StateVersionOutput{},
	}

	if from.LatestRun != nil {
		to.CurrentRun = &types.Run{ID: from.LatestRun.ID}
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: support is currently limited to a couple of resources.
	var opts []jsonapi.MarshalOption
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				unmarshaled, err := m.GetOrganization(r.Context(), from.Organization)
				if err != nil {
					return nil, nil, err
				}
				to.Organization = m.toOrganization(unmarshaled)
				opts = append(opts, jsonapi.MarshalInclude(to.Organization))
			case "outputs":
				sv, err := m.GetCurrentStateVersion(r.Context(), from.ID)
				if err != nil {
					return nil, nil, err
				}
				for _, out := range sv.Outputs {
					to.Outputs = append(to.Outputs, m.toOutput(out))
					opts = append(opts, jsonapi.MarshalInclude(m.toOutput(out)))
				}
			}
		}
	}

	return to, opts, nil
}

func (m *jsonapiMarshaler) toWorkspaceList(from *workspace.WorkspaceList, r *http.Request) (to []*types.Workspace, marshalOpts []jsonapi.MarshalOption, err error) {
	marshalOpts = []jsonapi.MarshalOption{toMarshalOption(from.Pagination)}
	for _, ws := range from.Items {
		item, itemOpts, err := m.toWorkspace(ws, r)
		if err != nil {
			return nil, nil, err
		}
		to = append(to, item)
		marshalOpts = append(marshalOpts, itemOpts...)
	}
	return to, marshalOpts, nil
}
