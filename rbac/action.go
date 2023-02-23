package rbac

// Action identifies an action a subject carries out on a resource for
// authorization purposes.
type Action int

const (
	WatchAction Action = iota
	CreateOrganizationAction
	UpdateOrganizationAction
	GetOrganizationAction
	GetEntitlementsAction
	DeleteOrganizationAction

	CreateVCSProviderAction
	GetVCSProviderAction
	ListVCSProvidersAction
	DeleteVCSProviderAction

	CreateAgentTokenAction
	ListAgentTokensAction
	DeleteAgentTokenAction

	CreateRegistrySessionAction

	CreateModuleAction
	CreateModuleVersionAction
	UpdateModuleAction
	ListModulesAction
	GetModuleAction
	DeleteModuleAction

	CreateVariableAction
	UpdateVariableAction
	ListVariablesAction
	GetVariableAction
	DeleteVariableAction

	GetRunAction
	ListRunsAction
	ApplyRunAction
	CreateRunAction
	DiscardRunAction
	DeleteRunAction
	CancelRunAction
	EnqueuePlanAction
	StartPhaseAction
	FinishPhaseAction
	PutChunkAction
	TailLogsAction

	GetPlanFileAction
	UploadPlanFileAction

	GetLockFileAction
	UploadLockFileAction

	ListWorkspacesAction
	GetWorkspaceAction
	CreateWorkspaceAction
	DeleteWorkspaceAction
	SetWorkspacePermissionAction
	UnsetWorkspacePermissionAction
	LockWorkspaceAction
	UnlockWorkspaceAction
	UpdateWorkspaceAction

	CreateStateVersionAction
	ListStateVersionsAction
	GetStateVersionAction
	DownloadStateAction
	GetStateVersionOutputAction

	CreateConfigurationVersionAction
	ListConfigurationVersionsAction
	GetConfigurationVersionAction
	DownloadConfigurationVersionAction

	ListUsersAction

	CreateTeamAction
	UpdateTeamAction
	GetTeamAction
	ListTeamsAction
)
