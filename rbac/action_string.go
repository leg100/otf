// Code generated by "stringer -type Action ./rbac"; DO NOT EDIT.

package rbac

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[WatchAction-0]
	_ = x[CreateOrganizationAction-1]
	_ = x[UpdateOrganizationAction-2]
	_ = x[GetOrganizationAction-3]
	_ = x[GetEntitlementsAction-4]
	_ = x[DeleteOrganizationAction-5]
	_ = x[CreateVCSProviderAction-6]
	_ = x[GetVCSProviderAction-7]
	_ = x[ListVCSProvidersAction-8]
	_ = x[DeleteVCSProviderAction-9]
	_ = x[CreateAgentTokenAction-10]
	_ = x[ListAgentTokensAction-11]
	_ = x[DeleteAgentTokenAction-12]
	_ = x[CreateRegistrySessionAction-13]
	_ = x[CreateModuleAction-14]
	_ = x[CreateModuleVersionAction-15]
	_ = x[UpdateModuleAction-16]
	_ = x[ListModulesAction-17]
	_ = x[GetModuleAction-18]
	_ = x[DeleteModuleAction-19]
	_ = x[CreateVariableAction-20]
	_ = x[UpdateVariableAction-21]
	_ = x[ListVariablesAction-22]
	_ = x[GetVariableAction-23]
	_ = x[DeleteVariableAction-24]
	_ = x[GetRunAction-25]
	_ = x[ListRunsAction-26]
	_ = x[ApplyRunAction-27]
	_ = x[CreateRunAction-28]
	_ = x[DiscardRunAction-29]
	_ = x[DeleteRunAction-30]
	_ = x[CancelRunAction-31]
	_ = x[EnqueuePlanAction-32]
	_ = x[StartPhaseAction-33]
	_ = x[FinishPhaseAction-34]
	_ = x[PutChunkAction-35]
	_ = x[TailLogsAction-36]
	_ = x[GetPlanFileAction-37]
	_ = x[UploadPlanFileAction-38]
	_ = x[GetLockFileAction-39]
	_ = x[UploadLockFileAction-40]
	_ = x[ListWorkspacesAction-41]
	_ = x[GetWorkspaceAction-42]
	_ = x[CreateWorkspaceAction-43]
	_ = x[DeleteWorkspaceAction-44]
	_ = x[SetWorkspacePermissionAction-45]
	_ = x[UnsetWorkspacePermissionAction-46]
	_ = x[UpdateWorkspaceAction-47]
	_ = x[LockWorkspaceAction-48]
	_ = x[UnlockWorkspaceAction-49]
	_ = x[ForceUnlockWorkspaceAction-50]
	_ = x[CreateStateVersionAction-51]
	_ = x[ListStateVersionsAction-52]
	_ = x[GetStateVersionAction-53]
	_ = x[DownloadStateAction-54]
	_ = x[CreateConfigurationVersionAction-55]
	_ = x[ListConfigurationVersionsAction-56]
	_ = x[GetConfigurationVersionAction-57]
	_ = x[DownloadConfigurationVersionAction-58]
	_ = x[ListUsersAction-59]
	_ = x[CreateTeamAction-60]
	_ = x[UpdateTeamAction-61]
	_ = x[GetTeamAction-62]
	_ = x[ListTeamsAction-63]
}

const _Action_name = "WatchActionCreateOrganizationActionUpdateOrganizationActionGetOrganizationActionGetEntitlementsActionDeleteOrganizationActionCreateVCSProviderActionGetVCSProviderActionListVCSProvidersActionDeleteVCSProviderActionCreateAgentTokenActionListAgentTokensActionDeleteAgentTokenActionCreateRegistrySessionActionCreateModuleActionCreateModuleVersionActionUpdateModuleActionListModulesActionGetModuleActionDeleteModuleActionCreateVariableActionUpdateVariableActionListVariablesActionGetVariableActionDeleteVariableActionGetRunActionListRunsActionApplyRunActionCreateRunActionDiscardRunActionDeleteRunActionCancelRunActionEnqueuePlanActionStartPhaseActionFinishPhaseActionPutChunkActionTailLogsActionGetPlanFileActionUploadPlanFileActionGetLockFileActionUploadLockFileActionListWorkspacesActionGetWorkspaceActionCreateWorkspaceActionDeleteWorkspaceActionSetWorkspacePermissionActionUnsetWorkspacePermissionActionUpdateWorkspaceActionLockWorkspaceActionUnlockWorkspaceActionForceUnlockWorkspaceActionCreateStateVersionActionListStateVersionsActionGetStateVersionActionDownloadStateActionCreateConfigurationVersionActionListConfigurationVersionsActionGetConfigurationVersionActionDownloadConfigurationVersionActionListUsersActionCreateTeamActionUpdateTeamActionGetTeamActionListTeamsAction"

var _Action_index = [...]uint16{0, 11, 35, 59, 80, 101, 125, 148, 168, 190, 213, 235, 256, 278, 305, 323, 348, 366, 383, 398, 416, 436, 456, 475, 492, 512, 524, 538, 552, 567, 583, 598, 613, 630, 646, 663, 677, 691, 708, 728, 745, 765, 785, 803, 824, 845, 873, 903, 924, 943, 964, 990, 1014, 1037, 1058, 1077, 1109, 1140, 1169, 1203, 1218, 1234, 1250, 1263, 1278}

func (i Action) String() string {
	if i < 0 || i >= Action(len(_Action_index)-1) {
		return "Action(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Action_name[_Action_index[i]:_Action_index[i+1]]
}
