// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

type Agent struct {
	AgentID      pgtype.Text
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IpAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  pgtype.Text
}

type AgentPool struct {
	AgentPoolID        pgtype.Text
	Name               pgtype.Text
	CreatedAt          pgtype.Timestamptz
	OrganizationName   pgtype.Text
	OrganizationScoped pgtype.Bool
}

type AgentPoolAllowedWorkspace struct {
	AgentPoolID pgtype.Text
	WorkspaceID pgtype.Text
}

type AgentStatus struct {
	Status pgtype.Text
}

type AgentToken struct {
	AgentTokenID pgtype.Text
	CreatedAt    pgtype.Timestamptz
	Description  pgtype.Text
	AgentPoolID  pgtype.Text
}

type Apply struct {
	RunID          pgtype.Text
	Status         pgtype.Text
	ResourceReport interface{}
}

type ConfigurationVersion struct {
	ConfigurationVersionID pgtype.Text
	CreatedAt              pgtype.Timestamptz
	AutoQueueRuns          pgtype.Bool
	Source                 pgtype.Text
	Speculative            pgtype.Bool
	Status                 pgtype.Text
	Config                 []byte
	WorkspaceID            pgtype.Text
}

type ConfigurationVersionIngressAttribute struct {
	Branch                 pgtype.Text
	CommitSha              pgtype.Text
	Identifier             pgtype.Text
	IsPullRequest          pgtype.Bool
	OnDefaultBranch        pgtype.Bool
	ConfigurationVersionID pgtype.Text
	CommitURL              pgtype.Text
	PullRequestNumber      pgtype.Int4
	PullRequestURL         pgtype.Text
	PullRequestTitle       pgtype.Text
	Tag                    pgtype.Text
	SenderUsername         pgtype.Text
	SenderAvatarURL        pgtype.Text
	SenderHtmlURL          pgtype.Text
}

type ConfigurationVersionStatusTimestamp struct {
	ConfigurationVersionID pgtype.Text
	Status                 pgtype.Text
	Timestamp              pgtype.Timestamptz
}

type DestinationType struct {
	Name pgtype.Text
}

type GithubApp struct {
	GithubAppID   pgtype.Int8
	WebhookSecret pgtype.Text
	PrivateKey    pgtype.Text
	Slug          pgtype.Text
	Organization  pgtype.Text
}

type GithubAppInstall struct {
	GithubAppID   pgtype.Int8
	InstallID     pgtype.Int8
	Username      pgtype.Text
	Organization  pgtype.Text
	VCSProviderID pgtype.Text
}

type IngressAttribute struct {
	Branch                 pgtype.Text
	CommitSha              pgtype.Text
	Identifier             pgtype.Text
	IsPullRequest          pgtype.Bool
	OnDefaultBranch        pgtype.Bool
	ConfigurationVersionID pgtype.Text
	CommitURL              pgtype.Text
	PullRequestNumber      pgtype.Int4
	PullRequestURL         pgtype.Text
	PullRequestTitle       pgtype.Text
	Tag                    pgtype.Text
	SenderUsername         pgtype.Text
	SenderAvatarURL        pgtype.Text
	SenderHtmlURL          pgtype.Text
}

type Job struct {
	RunID    pgtype.Text
	Phase    pgtype.Text
	Status   pgtype.Text
	AgentID  pgtype.Text
	Signaled pgtype.Bool
}

type JobPhase struct {
	Phase pgtype.Text
}

type JobStatus struct {
	Status pgtype.Text
}

type LatestTerraformVersion struct {
	Version    pgtype.Text
	Checkpoint pgtype.Timestamptz
}

type Log struct {
	RunID   pgtype.Text
	Phase   pgtype.Text
	ChunkID pgtype.Int4
	Chunk   []byte
	Offset  pgtype.Int4
}

type Module struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
}

type ModuleStatus struct {
	Status pgtype.Text
}

type ModuleTarball struct {
	Tarball         []byte
	ModuleVersionID pgtype.Text
}

type ModuleVersion struct {
	ModuleVersionID pgtype.Text
	Version         pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	Status          pgtype.Text
	StatusError     pgtype.Text
	ModuleID        pgtype.Text
}

type ModuleVersionStatus struct {
	Status pgtype.Text
}

type NotificationConfiguration struct {
	NotificationConfigurationID pgtype.Text
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	Name                        pgtype.Text
	URL                         pgtype.Text
	Triggers                    []pgtype.Text
	DestinationType             pgtype.Text
	WorkspaceID                 pgtype.Text
	Enabled                     pgtype.Bool
}

type Organization struct {
	OrganizationID             pgtype.Text
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	Name                       pgtype.Text
	SessionRemember            pgtype.Int4
	SessionTimeout             pgtype.Int4
	Email                      pgtype.Text
	CollaboratorAuthPolicy     pgtype.Text
	AllowForceDeleteWorkspaces pgtype.Bool
	CostEstimationEnabled      pgtype.Bool
}

type OrganizationToken struct {
	OrganizationTokenID pgtype.Text
	CreatedAt           pgtype.Timestamptz
	OrganizationName    pgtype.Text
	Expiry              pgtype.Timestamptz
}

type Phase struct {
	Phase pgtype.Text
}

type PhaseStatus struct {
	Status pgtype.Text
}

type PhaseStatusTimestamp struct {
	RunID     pgtype.Text
	Phase     pgtype.Text
	Status    pgtype.Text
	Timestamp pgtype.Timestamptz
}

type Plan struct {
	RunID          pgtype.Text
	Status         pgtype.Text
	PlanBin        []byte
	PlanJson       []byte
	ResourceReport interface{}
	OutputReport   interface{}
}

type RegistrySession struct {
	Token            pgtype.Text
	Expiry           pgtype.Timestamptz
	OrganizationName pgtype.Text
}

type RepoConnection struct {
	ModuleID      pgtype.Text
	WorkspaceID   pgtype.Text
	RepoPath      pgtype.Text
	VCSProviderID pgtype.Text
}

type Repohook struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSProviderID pgtype.Text
}

type Run struct {
	RunID                  pgtype.Text
	CreatedAt              pgtype.Timestamptz
	CancelSignaledAt       pgtype.Timestamptz
	IsDestroy              pgtype.Bool
	PositionInQueue        pgtype.Int4
	Refresh                pgtype.Bool
	RefreshOnly            pgtype.Bool
	ReplaceAddrs           []pgtype.Text
	TargetAddrs            []pgtype.Text
	LockFile               []byte
	Status                 pgtype.Text
	WorkspaceID            pgtype.Text
	ConfigurationVersionID pgtype.Text
	AutoApply              pgtype.Bool
	PlanOnly               pgtype.Bool
	CreatedBy              pgtype.Text
	Source                 pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
}

type RunStatus struct {
	Status pgtype.Text
}

type RunStatusTimestamp struct {
	RunID     pgtype.Text
	Status    pgtype.Text
	Timestamp pgtype.Timestamptz
}

type RunVariable struct {
	RunID pgtype.Text
	Key   pgtype.Text
	Value pgtype.Text
}

type SchemaVersion struct {
	Version pgtype.Int4
}

type Session struct {
	Token     pgtype.Text
	CreatedAt pgtype.Timestamptz
	Address   pgtype.Text
	Expiry    pgtype.Timestamptz
	Username  pgtype.Text
}

type StateVersion struct {
	StateVersionID pgtype.Text
	CreatedAt      pgtype.Timestamptz
	Serial         pgtype.Int4
	State          []byte
	WorkspaceID    pgtype.Text
	Status         pgtype.Text
}

type StateVersionOutput struct {
	StateVersionOutputID pgtype.Text
	Name                 pgtype.Text
	Sensitive            pgtype.Bool
	Type                 pgtype.Text
	Value                []byte
	StateVersionID       pgtype.Text
}

type StateVersionStatus struct {
	Status pgtype.Text
}

type Tag struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
}

type Team struct {
	TeamID                          pgtype.Text
	Name                            pgtype.Text
	CreatedAt                       pgtype.Timestamptz
	PermissionManageWorkspaces      pgtype.Bool
	PermissionManageVCS             pgtype.Bool
	PermissionManageModules         pgtype.Bool
	OrganizationName                pgtype.Text
	SSOTeamID                       pgtype.Text
	Visibility                      pgtype.Text
	PermissionManagePolicies        pgtype.Bool
	PermissionManagePolicyOverrides pgtype.Bool
	PermissionManageProviders       pgtype.Bool
}

type TeamMembership struct {
	TeamID   pgtype.Text
	Username pgtype.Text
}

type TeamToken struct {
	TeamTokenID pgtype.Text
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	TeamID      pgtype.Text
	Expiry      pgtype.Timestamptz
}

type Token struct {
	TokenID     pgtype.Text
	CreatedAt   pgtype.Timestamptz
	Description pgtype.Text
	Username    pgtype.Text
}

type User struct {
	UserID    pgtype.Text
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
}

type VCSKind struct {
	Name pgtype.Text
}

type VCSProvider struct {
	VCSProviderID    pgtype.Text
	Token            pgtype.Text
	CreatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	VCSKind          pgtype.Text
	OrganizationName pgtype.Text
	GithubAppID      pgtype.Int8
}

type Variable struct {
	VariableID  pgtype.Text
	Key         pgtype.Text
	Value       pgtype.Text
	Description pgtype.Text
	Category    pgtype.Text
	Sensitive   pgtype.Bool
	Hcl         pgtype.Bool
	VersionID   pgtype.Text
}

type VariableCategory struct {
	Category pgtype.Text
}

type VariableSet struct {
	VariableSetID    pgtype.Text
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
}

type VariableSetVariable struct {
	VariableSetID pgtype.Text
	VariableID    pgtype.Text
}

type VariableSetWorkspace struct {
	VariableSetID pgtype.Text
	WorkspaceID   pgtype.Text
}

type Workspace struct {
	WorkspaceID                pgtype.Text
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	AllowDestroyPlan           pgtype.Bool
	AutoApply                  pgtype.Bool
	CanQueueDestroyPlan        pgtype.Bool
	Description                pgtype.Text
	Environment                pgtype.Text
	ExecutionMode              pgtype.Text
	GlobalRemoteState          pgtype.Bool
	MigrationEnvironment       pgtype.Text
	Name                       pgtype.Text
	QueueAllRuns               pgtype.Bool
	SpeculativeEnabled         pgtype.Bool
	SourceName                 pgtype.Text
	SourceURL                  pgtype.Text
	StructuredRunOutputEnabled pgtype.Bool
	TerraformVersion           pgtype.Text
	TriggerPrefixes            []pgtype.Text
	WorkingDirectory           pgtype.Text
	LockRunID                  pgtype.Text
	LatestRunID                pgtype.Text
	OrganizationName           pgtype.Text
	Branch                     pgtype.Text
	LockUsername               pgtype.Text
	CurrentStateVersionID      pgtype.Text
	TriggerPatterns            []pgtype.Text
	VCSTagsRegex               pgtype.Text
	AllowCLIApply              pgtype.Bool
	AgentPoolID                pgtype.Text
}

type WorkspacePermission struct {
	WorkspaceID pgtype.Text
	TeamID      pgtype.Text
	Role        pgtype.Text
}

type WorkspaceRole struct {
	Role pgtype.Text
}

type WorkspaceTag struct {
	TagID       pgtype.Text
	WorkspaceID pgtype.Text
}

type WorkspaceVariable struct {
	WorkspaceID pgtype.Text
	VariableID  pgtype.Text
}
