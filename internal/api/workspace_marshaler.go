package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/workspace"
)

func (m *jsonapiMarshaler) toWorkspace(from *workspace.Workspace, r *http.Request) (*types.Workspace, []jsonapi.MarshalOption, error) {
	subject, err := internal.SubjectFromContext(r.Context())
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
		TriggerPatterns:            from.TriggerPatterns,
		WorkingDirectory:           from.WorkingDirectory,
		TagNames:                   from.Tags,
		UpdatedAt:                  from.UpdatedAt,
		Organization:               &types.Organization{Name: from.Organization},
		Outputs:                    []*types.StateVersionOutput{},
	}
	if len(from.TriggerPrefixes) > 0 || len(from.TriggerPatterns) > 0 {
		to.FileTriggersEnabled = true
	}
	if from.LatestRun != nil {
		to.CurrentRun = &types.Run{ID: from.LatestRun.ID}
	}
	if from.Connection != nil {
		to.VCSRepo = &types.VCSRepo{
			OAuthTokenID: from.Connection.VCSProviderID,
			Branch:       from.Connection.Branch,
			Identifier:   from.Connection.Repo,
			TagsRegex:    from.Connection.TagsRegex,
		}
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
