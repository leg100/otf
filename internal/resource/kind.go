package resource

type Kind string

func (k Kind) String() string {
	return string(k)
}

const (
	SiteKind                      Kind = "site"
	OrganizationKind              Kind = "org"
	WorkspaceKind                 Kind = "ws"
	RunKind                       Kind = "run"
	ConfigVersionKind             Kind = "cv"
	IngressAttributesKind         Kind = "ia"
	JobKind                       Kind = "job"
	ChunkKind                     Kind = "chunk"
	UserKind                      Kind = "user"
	TeamKind                      Kind = "team"
	ModuleKind                    Kind = "mod"
	ModuleVersionKind             Kind = "modver"
	NotificationConfigurationKind Kind = "nc"
	AgentPoolKind                 Kind = "apool"
	RunnerKind                    Kind = "runner"
	StateVersionKind              Kind = "sv"
	StateVersionOutputKind        Kind = "wsout"
	VariableSetKind               Kind = "varset"
	VariableKind                  Kind = "var"
	VCSProviderKind               Kind = "vcs"
	OrganizationTokenKind         Kind = "ot"
	UserTokenKind                 Kind = "ut"
	TeamTokenKind                 Kind = "tt"
	AgentTokenKind                Kind = "at"
	SSHKeyKind                    Kind = "sshkey"
	RunTriggerKind                Kind = "rt"
)

var fullKinds = map[Kind]string{
	OrganizationKind:              "organization",
	WorkspaceKind:                 "workspace",
	RunKind:                       "run",
	ConfigVersionKind:             "config",
	IngressAttributesKind:         "ingress-attributes",
	JobKind:                       "job",
	ModuleKind:                    "module",
	ModuleVersionKind:             "module-version",
	NotificationConfigurationKind: "notification-config",
	AgentPoolKind:                 "agent-pool",
	RunnerKind:                    "runner",
	StateVersionKind:              "state-version",
	StateVersionOutputKind:        "workspace-output",
	VariableSetKind:               "variable-set",
	VariableKind:                  "variable",
	OrganizationTokenKind:         "organization-token",
	UserTokenKind:                 "user-token",
	TeamTokenKind:                 "team-token",
	AgentTokenKind:                "agent-token",
	SSHKeyKind:                    "ssh-key",
	RunTriggerKind:                "trigger",
}

// Full returns the unabbreviated name for the kind.
func (k Kind) Full() string {
	if full, ok := fullKinds[k]; ok {
		return full
	}
	return k.String()
}
