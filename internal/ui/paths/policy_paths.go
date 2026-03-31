package paths

import "fmt"

func PolicySets(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets", organization)
}

func NewManualPolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/connect/manual", organization)
}

func CreatePolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/create", organization)
}

func ConnectPolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/connect", organization)
}

func ConnectVCSPolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/connect/vcs", organization)
}

func PreviewVCSPolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/connect/vcs/preview", organization)
}

func CreateVCSPolicySet(organization any) string {
	return fmt.Sprintf("/app/organizations/%v/policy-sets/connect/vcs/create", organization)
}

func SyncPolicySet(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v/sync", policySet)
}

func PolicySet(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v", policySet)
}

func UpdatePolicySet(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v/update", policySet)
}

func DeletePolicySet(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v/delete", policySet)
}

func CreatePolicy(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v/policies/create", policySet)
}

func UpdatePolicy(policy any) string {
	return fmt.Sprintf("/app/policies/%v/update", policy)
}

func DeletePolicy(policy any) string {
	return fmt.Sprintf("/app/policies/%v/delete", policy)
}

func UpdatePolicySetWorkspaces(policySet any) string {
	return fmt.Sprintf("/app/policy-sets/%v/workspaces/update", policySet)
}

func WorkspaceSentinel(workspace any) string {
	return fmt.Sprintf("/app/workspaces/%v/sentinel", workspace)
}

func DownloadWorkspaceMocks(workspace any) string {
	return fmt.Sprintf("/app/workspaces/%v/sentinel/mocks", workspace)
}
