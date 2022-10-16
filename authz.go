package otf

import (
	"context"
	"errors"
	"fmt"
)

// unexported key type prevents collisions
type subjectCtxKeyType string

const subjectCtxKey subjectCtxKeyType = "subject"

var ErrAccessNotPermitted = errors.New("access to the resource is not permitted")

// Subject is an entity attempting to carry out an action on a resource.
type Subject interface {
	CanAccessSite(action Action) bool
	CanAccessOrganization(action Action, name string) bool
	CanAccessWorkspace(action Action, policy *WorkspacePolicy) bool
	Identity
}

type Action string

const (
	WatchAction Action = "watch"

	CreateOrganizationAction Action = "create_organization"
	UpdateOrganizationAction Action = "update_organization"
	GetOrganizationAction    Action = "get_organization"
	GetEntitlementsAction    Action = "get_entitlements"
	DeleteOrganizationAction Action = "delete_organization"

	CreateAgentTokenAction Action = "create_agent_token"
	ListAgentTokenActions  Action = "list_agent_tokens"
	DeleteAgentTokenAction Action = "delete_agent_token"

	GetRunAction      Action = "get_run"
	ListRunsAction    Action = "list_runs"
	ApplyRunAction    Action = "apply_run"
	CreateRunAction   Action = "create_run"
	DiscardRunAction  Action = "discard_run"
	DeleteRunAction   Action = "delete_run"
	CancelRunAction   Action = "cancel_run"
	EnqueuePlanAction Action = "enqueue_plan"
	StartPhaseAction  Action = "start_run_phase"
	FinishPhaseAction Action = "finish_run_phase"
	PutChunkAction    Action = "put_log_chunk"
	TailLogsAction    Action = "tail_logs"

	GetPlanFileAction    Action = "get_plan_file"
	UploadPlanFileAction Action = "upload_plan_file"

	GetLockFileAction    Action = "get_lock_file"
	UploadLockFileAction Action = "upload_lock_file"

	ListWorkspacesAction         Action = "list_workspaces"
	GetWorkspaceAction           Action = "get_workspace"
	CreateWorkspaceAction        Action = "create_workspace"
	DeleteWorkspaceAction        Action = "delete_workspace"
	SetWorkspacePermissionAction Action = "set_workspace_permission"

	CreateStateVersionAction Action = "create_state_version"
	ListStateVersionsAction  Action = "list_state_versions"
	GetStateVersionAction    Action = "get_state_version"
	DownloadStateAction      Action = "download_state"

	ListUsersAction Action = "list_users"
)

var workspaceManagerPermissions = map[Action]bool{
	CreateWorkspaceAction: true,
}

var adminPermissions = map[Action]bool{
	SetWorkspacePermissionAction: true,
	DeleteWorkspaceAction:        true,
}

var writePermissions = map[Action]bool{
	ApplyRunAction: true,
}

var planPermissions = map[Action]bool{
	CreateRunAction: true,
}

var readPermissions = map[Action]bool{
	ListRunsAction:    true,
	GetPlanFileAction: true,
}

func init() {
	// plan role includes read permissions
	for p := range readPermissions {
		planPermissions[p] = true
	}
	// write role includes plan permissions
	for p := range planPermissions {
		writePermissions[p] = true
	}
	// admin role includes write permissions
	for p := range writePermissions {
		adminPermissions[p] = true
	}
	// workspace manager role includes admin permissions
	for p := range adminPermissions {
		workspaceManagerPermissions[p] = true
	}
}

type WorkspacePolicy struct {
	OrganizationName string
	WorkspaceID      string
	Permissions      []*WorkspacePermission
}

// AddSubjectToContext adds a subject to a context
func AddSubjectToContext(ctx context.Context, subj Subject) context.Context {
	return context.WithValue(ctx, subjectCtxKey, subj)
}

// SubjectFromContext retrieves a subject from a context
func SubjectFromContext(ctx context.Context) (Subject, error) {
	subj, ok := ctx.Value(subjectCtxKey).(Subject)
	if !ok {
		return nil, fmt.Errorf("no subject in context")
	}
	return subj, nil
}

// UserFromContext retrieves a user from a context
func UserFromContext(ctx context.Context) (*User, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subj.(*User)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not a user")
	}
	return user, nil
}

// AgentFromContext retrieves an agent(-token) from a context
func AgentFromContext(ctx context.Context) (*AgentToken, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*AgentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

// LockFromContext retrieves a workspace lock from a context
func LockFromContext(ctx context.Context) (WorkspaceLockState, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lock, ok := subj.(WorkspaceLockState)
	if !ok {
		return nil, fmt.Errorf("no lock subject in context")
	}
	return lock, nil
}

// IsAdmin determines if the caller is an admin, i.e. the app/agent/site-admin,
// but not a normal user. Returns false if the context contains no subject.
func IsAdmin(ctx context.Context) bool {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		// unauthenticated call
		return false
	}
	if user, ok := subj.(*User); ok && !user.IsSiteAdmin() {
		// is normal user
		return false
	}
	// call is authenticated and the subject is not a normal user
	return true
}

func IsAllowed(action Action, role WorkspaceRole) bool {
	switch role {
	case WorkspaceAdminRole:
		return true
	case WorkspaceWriteRole:
		return writePermissions[action]
	case WorkspacePlanRole:
		return planPermissions[action]
	case WorkspaceReadRole:
		return readPermissions[action]
	}
	return false
}
