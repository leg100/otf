package sql

// TODO: populate this list at init time: use observer pattern where each domain
// package is responsible for 'registering' its tables.
var tableTypes = []string{
	"ingress_attributes",
	"ingress_attributes[]",
	"workspace_permissions",
	"workspace_permissions[]",
	"report",
	"report[]",
	"run_variables",
	"run_variables[]",
	"run_status_timestamps",
	"run_status_timestamps[]",
	"phase_status_timestamps",
	"phase_status_timestamps[]",
	"configuration_version_status_timestamps",
	"configuration_version_status_timestamps[]",
	"teams",
	"teams[]",
	"github_apps",
	"github_apps[]",
	"agent_pools",
	"agent_pools[]",
	"state_version_outputs",
	"state_version_outputs[]",
	"variables",
	"variables[]",
	"module_versions",
	"module_versions[]",
	"repo_connections",
	"repo_connections[]",
}
