package otf

const (
	WatchAction Action = "watch"

	CreateOrganizationAction Action = "create_organization"
	UpdateOrganizationAction Action = "update_organization"
	GetOrganizationAction    Action = "get_organization"
	GetEntitlementsAction    Action = "get_entitlements"
	DeleteOrganizationAction Action = "delete_organization"

	CreateVCSProviderAction Action = "create_vcs_provider"
	GetVCSProviderAction    Action = "get_vcs_provider"
	ListVCSProvidersAction  Action = "list_vcs_provider"
	DeleteVCSProviderAction Action = "delete_vcs_provider"

	CreateAgentTokenAction Action = "create_agent_token"
	ListAgentTokensAction  Action = "list_agent_tokens"
	DeleteAgentTokenAction Action = "delete_agent_token"

	CreateRegistrySessionAction Action = "create_registry_session_token"

	CreateModuleAction        Action = "create_module"
	CreateModuleVersionAction Action = "create_module_version"
	UpdateModuleAction        Action = "update_module"
	ListModulesAction         Action = "list_modules"
	GetModuleAction           Action = "get_module"
	DeleteModuleAction        Action = "delete_modules"

	CreateVariableAction Action = "create_variable"
	UpdateVariableAction Action = "update_variable"
	ListVariablesAction  Action = "list_variables"
	GetVariableAction    Action = "get_variable"
	DeleteVariableAction Action = "delete_variables"

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

	ListWorkspacesAction           Action = "list_workspaces"
	GetWorkspaceAction             Action = "get_workspace"
	CreateWorkspaceAction          Action = "create_workspace"
	DeleteWorkspaceAction          Action = "delete_workspace"
	SetWorkspacePermissionAction   Action = "set_workspace_permission"
	UnsetWorkspacePermissionAction Action = "unset_workspace_permission"
	LockWorkspaceAction            Action = "lock_workspace"
	UnlockWorkspaceAction          Action = "unlock_workspace"
	UpdateWorkspaceAction          Action = "update_workspace"

	CreateStateVersionAction Action = "create_state_version"
	ListStateVersionsAction  Action = "list_state_versions"
	GetStateVersionAction    Action = "get_state_version"
	DownloadStateAction      Action = "download_state"

	CreateConfigurationVersionAction   Action = "create_configuration_version"
	ListConfigurationVersionsAction    Action = "list_configuration_versions"
	GetConfigurationVersionAction      Action = "get_configuration_version"
	DownloadConfigurationVersionAction Action = "download_configuration_version"

	ListUsersAction Action = "list_users"

	CreateTeamAction Action = "create_team"
	UpdateTeamAction Action = "update_team"
	GetTeamAction    Action = "get_team"
	ListTeamsAction  Action = "list_teams"
)

// Action symbolizes an action performed on a resource.
type Action string
