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
	"org":    "organization",
	"ws":     "workspace",
	"run":    "run",
	"cv":     "config",
	"ia":     "ingress-attributes",
	"job":    "job",
	"mod":    "module",
	"modver": "module-version",
	"nc":     "notification-config",
	"apool":  "agent-pool",
	"runner": "runner",
	"sv":     "state-version",
	"wsout":  "workspace-output",
	"varset": "variable-set",
	"var":    "variable",
	"ot":     "organization-token",
	"ut":     "user-token",
	"tt":     "team-token",
	"at":     "agent-token",
	"sshkey": "ssh-key",
	"rt":     "trigger",
}

// Full returns the unabbreviated name for the kind.
func (k Kind) Full() string {
	if full, ok := fullKinds[k]; ok {
		return full
	}
	return k.String()
}
