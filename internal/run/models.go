// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package run

import (
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

type AgentPool struct {
	AgentPoolID        resource.ID
	Name               pgtype.Text
	CreatedAt          pgtype.Timestamptz
	OrganizationName   pgtype.Text
	OrganizationScoped pgtype.Bool
}

type AgentPoolAllowedWorkspace struct {
	AgentPoolID resource.ID
	WorkspaceID resource.ID
}

type AgentToken struct {
	AgentTokenID resource.ID
	CreatedAt    pgtype.Timestamptz
	Description  pgtype.Text
	AgentPoolID  resource.ID
}

type Apply struct {
	RunID          resource.ID
	Status         pgtype.Text
	ResourceReport interface{}
}

type ConfigurationVersion struct {
	ConfigurationVersionID resource.ID
	CreatedAt              pgtype.Timestamptz
	AutoQueueRuns          pgtype.Bool
	Source                 pgtype.Text
	Speculative            pgtype.Bool
	Status                 pgtype.Text
	Config                 []byte
	WorkspaceID            resource.ID
}

type ConfigurationVersionStatusTimestamp struct {
	ConfigurationVersionID resource.ID
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
	VCSProviderID resource.ID
}

type IngressAttribute struct {
	Branch                 pgtype.Text
	CommitSHA              pgtype.Text
	Identifier             pgtype.Text
	IsPullRequest          pgtype.Bool
	OnDefaultBranch        pgtype.Bool
	ConfigurationVersionID resource.ID
	CommitURL              pgtype.Text
	PullRequestNumber      pgtype.Int4
	PullRequestURL         pgtype.Text
	PullRequestTitle       pgtype.Text
	Tag                    pgtype.Text
	SenderUsername         pgtype.Text
	SenderAvatarURL        pgtype.Text
	SenderHTMLURL          pgtype.Text
}

type Job struct {
	RunID      resource.ID
	PhaseModel pgtype.Text
	Status     pgtype.Text
	RunnerID   *resource.ID
	Signaled   pgtype.Bool
	JobID      resource.ID
}

type JobPhase struct {
	PhaseModel pgtype.Text
}

type JobStatus struct {
	Status pgtype.Text
}

type LatestTerraformVersion struct {
	Version    pgtype.Text
	Checkpoint pgtype.Timestamptz
}

type Log struct {
	RunID      resource.ID
	PhaseModel pgtype.Text
	Chunk      []byte
	Offset     pgtype.Int4
	ChunkID    resource.ID
}

type Model struct {
	RunID                  resource.ID
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
	WorkspaceID            resource.ID
	ConfigurationVersionID resource.ID
	AutoApply              pgtype.Bool
	PlanOnly               pgtype.Bool
	CreatedBy              pgtype.Text
	Source                 pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
}

type Module struct {
	ModuleID         resource.ID
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
	ModuleVersionID resource.ID
}

type ModuleVersion struct {
	ModuleVersionID resource.ID
	Version         pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	Status          pgtype.Text
	StatusError     pgtype.Text
	ModuleID        resource.ID
}

type ModuleVersionStatus struct {
	Status pgtype.Text
}

type NotificationConfiguration struct {
	NotificationConfigurationID resource.ID
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	Name                        pgtype.Text
	URL                         pgtype.Text
	Triggers                    []pgtype.Text
	DestinationType             pgtype.Text
	WorkspaceID                 resource.ID
	Enabled                     pgtype.Bool
}

type Organization struct {
	OrganizationID             resource.ID
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
	OrganizationTokenID resource.ID
	CreatedAt           pgtype.Timestamptz
	OrganizationName    pgtype.Text
	Expiry              pgtype.Timestamptz
}

type PhaseModel struct {
	PhaseModel pgtype.Text
}

type PhaseStatusModel struct {
	Status pgtype.Text
}

type PhaseStatusTimestampModel struct {
	RunID      resource.ID
	PhaseModel pgtype.Text
	Status     pgtype.Text
	Timestamp  pgtype.Timestamptz
}

type Plan struct {
	RunID          resource.ID
	Status         pgtype.Text
	PlanBin        []byte
	PlanJSON       []byte
	ResourceReport interface{}
	OutputReport   interface{}
}

type RegistrySession struct {
	Token            pgtype.Text
	Expiry           pgtype.Timestamptz
	OrganizationName pgtype.Text
}

type RepoConnection struct {
	ModuleID      *resource.ID
	WorkspaceID   *resource.ID
	RepoPath      pgtype.Text
	VCSProviderID resource.ID
}

type Repohook struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSProviderID resource.ID
}

type RunStatus struct {
	Status pgtype.Text
}

type RunStatusTimestamp struct {
	RunID     resource.ID
	Status    pgtype.Text
	Timestamp pgtype.Timestamptz
}

type RunVariable struct {
	RunID resource.ID
	Key   pgtype.Text
	Value pgtype.Text
}

type Runner struct {
	RunnerID     resource.ID
	Name         pgtype.Text
	Version      pgtype.Text
	MaxJobs      pgtype.Int4
	IPAddress    netip.Addr
	LastPingAt   pgtype.Timestamptz
	LastStatusAt pgtype.Timestamptz
	Status       pgtype.Text
	AgentPoolID  *resource.ID
}

type RunnerStatus struct {
	Status pgtype.Text
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
	StateVersionID resource.ID
	CreatedAt      pgtype.Timestamptz
	Serial         pgtype.Int4
	State          []byte
	WorkspaceID    resource.ID
	Status         pgtype.Text
}

type StateVersionOutput struct {
	StateVersionOutputID resource.ID
	Name                 pgtype.Text
	Sensitive            pgtype.Bool
	Type                 pgtype.Text
	Value                []byte
	StateVersionID       resource.ID
}

type StateVersionStatus struct {
	Status pgtype.Text
}

type Tag struct {
	TagID            resource.ID
	Name             pgtype.Text
	OrganizationName pgtype.Text
}

type Team struct {
	TeamID                          resource.ID
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
	TeamID   resource.ID
	Username pgtype.Text
}

type TeamToken struct {
	TeamTokenID resource.ID
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	TeamID      resource.ID
	Expiry      pgtype.Timestamptz
}

type Token struct {
	TokenID     resource.ID
	CreatedAt   pgtype.Timestamptz
	Description pgtype.Text
	Username    pgtype.Text
}

type User struct {
	UserID    resource.ID
	Username  pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	SiteAdmin pgtype.Bool
}

type VCSKind struct {
	Name pgtype.Text
}

type VCSProvider struct {
	VCSProviderID    resource.ID
	Token            pgtype.Text
	CreatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	VCSKind          pgtype.Text
	OrganizationName pgtype.Text
	GithubAppID      pgtype.Int8
}

type VariableCategory struct {
	Category pgtype.Text
}

type VariableModel struct {
	VariableID  resource.ID
	Key         pgtype.Text
	Value       pgtype.Text
	Description pgtype.Text
	Category    pgtype.Text
	Sensitive   pgtype.Bool
	HCL         pgtype.Bool
	VersionID   pgtype.Text
}

type VariableSet struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
}

type VariableSetVariable struct {
	VariableSetID resource.ID
	VariableID    resource.ID
}

type VariableSetWorkspace struct {
	VariableSetID resource.ID
	WorkspaceID   resource.ID
}

type Workspace struct {
	WorkspaceID                resource.ID
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
	LockRunID                  *resource.ID
	LatestRunID                *resource.ID
	OrganizationName           pgtype.Text
	Branch                     pgtype.Text
	CurrentStateVersionID      *resource.ID
	TriggerPatterns            []pgtype.Text
	VCSTagsRegex               pgtype.Text
	AllowCLIApply              pgtype.Bool
	AgentPoolID                *resource.ID
	LockUserID                 *resource.ID
}

type WorkspacePermission struct {
	WorkspaceID resource.ID
	TeamID      resource.ID
	Role        pgtype.Text
}

type WorkspaceRole struct {
	Role pgtype.Text
}

type WorkspaceTag struct {
	TagID       resource.ID
	WorkspaceID resource.ID
}

type WorkspaceVariable struct {
	WorkspaceID resource.ID
	VariableID  resource.ID
}
