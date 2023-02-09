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
	r *http.Request
	otf.Application
	*Workspace
}

func (m *JSONAPIMarshaler) ToJSONAPI() any {
	subject, err := otf.SubjectFromContext(m.r.Context())
	if err != nil {
		panic(err.Error())
	}
	perms, err := m.ListWorkspacePermissions(m.r.Context(), m.ID())
	if err != nil {
		panic(err.Error())
	}
	policy := &otf.WorkspacePolicy{
		Organization: m.Organization(),
		WorkspaceID:  m.ID(),
		Permissions:  perms,
	}

	obj := &jsonapi.Workspace{
		ID: m.ID(),
		Actions: &jsonapi.WorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     m.AllowDestroyPlan(),
		AutoApply:            m.AutoApply(),
		CanQueueDestroyPlan:  m.CanQueueDestroyPlan(),
		CreatedAt:            m.CreatedAt(),
		Description:          m.Description(),
		Environment:          m.Environment(),
		ExecutionMode:        string(m.ExecutionMode()),
		FileTriggersEnabled:  m.FileTriggersEnabled(),
		GlobalRemoteState:    m.GlobalRemoteState(),
		Locked:               m.Locked(),
		MigrationEnvironment: m.MigrationEnvironment(),
		Name:                 m.Name(),
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations: m.ExecutionMode() == "remote",
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
		QueueAllRuns:               m.QueueAllRuns(),
		SpeculativeEnabled:         m.SpeculativeEnabled(),
		SourceName:                 m.SourceName(),
		SourceURL:                  m.SourceURL(),
		StructuredRunOutputEnabled: m.StructuredRunOutputEnabled(),
		TerraformVersion:           m.TerraformVersion(),
		TriggerPrefixes:            m.TriggerPrefixes(),
		WorkingDirectory:           m.WorkingDirectory(),
		UpdatedAt:                  m.UpdatedAt(),
		Organization:               &jsonapi.Organization{Name: m.Organization()},
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	//
	// NOTE: limit support to organization, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := m.r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "organization":
				org, err := m.GetOrganization(m.r.Context(), m.Organization())
				if err != nil {
					panic(err.Error()) // throws HTTP500
				}
				obj.Organization = (&Organization{org}).ToJSONAPI().(*jsonapi.Organization)
			}
		}
	}
	return obj
}
