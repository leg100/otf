// Package workspace provides access to terraform workspaces
package workspace

import (
	"errors"
	"time"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/semver"
	"golang.org/x/exp/slog"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true

	MinTerraformVersion     = "1.2.0"
	DefaultTerraformVersion = "1.3.7"
)

type (
	// Workspace is a terraform workspace.
	Workspace struct {
		ID                         string
		CreatedAt                  time.Time
		UpdatedAt                  time.Time
		AllowDestroyPlan           bool
		AutoApply                  bool
		Branch                     string
		CanQueueDestroyPlan        bool
		Description                string
		Environment                string
		ExecutionMode              ExecutionMode
		FileTriggersEnabled        bool
		GlobalRemoteState          bool
		MigrationEnvironment       string
		Name                       string
		QueueAllRuns               bool
		SpeculativeEnabled         bool
		StructuredRunOutputEnabled bool
		SourceName                 string
		SourceURL                  string
		TerraformVersion           string
		TriggerPrefixes            []string
		WorkingDirectory           string
		Organization               string
		Connection                 *repo.Connection
		LatestRun                  *LatestRun
		Tags                       []string

		*lock
	}

	// LatestRun is a summary of the latest run for a workspace
	LatestRun struct {
		ID     string
		Status internal.RunStatus
	}

	ExecutionMode string

	// WorkspaceList is a list of workspaces.
	WorkspaceList struct {
		*internal.Pagination
		Items []*Workspace
	}

	// CreateOptions represents the options for creating a new workspace.
	CreateOptions struct {
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Branch                     *string
		Description                *string
		ExecutionMode              *ExecutionMode
		FileTriggersEnabled        *bool
		GlobalRemoteState          *bool
		MigrationEnvironment       *string
		Name                       *string `schema:"name,required"`
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		SourceName                 *string
		SourceURL                  *string
		StructuredRunOutputEnabled *bool
		Tags                       []TagSpec
		TerraformVersion           *string
		TriggerPrefixes            []string
		WorkingDirectory           *string
		Organization               *string `schema:"organization_name,required"`

		*ConnectOptions
	}

	UpdateOptions struct {
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Name                       *string
		Description                *string
		ExecutionMode              *ExecutionMode `schema:"execution_mode"`
		FileTriggersEnabled        *bool
		GlobalRemoteState          *bool
		Operations                 *bool
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		StructuredRunOutputEnabled *bool
		TerraformVersion           *string `schema:"terraform_version"`
		TriggerPrefixes            []string
		WorkingDirectory           *string
	}

	// ListOptions are options for paginating and filtering a list of
	// Workspaces
	ListOptions struct {
		internal.ListOptions          // Pagination
		Prefix               string   `schema:"search[name],omitempty"`
		Tags                 []string `schema:"search[tags],omitempty"`
		Organization         *string  `schema:"organization_name,required"`
	}

	ConnectOptions struct {
		RepoPath      string `schema:"identifier,required"` // repo id: <owner>/<repo>
		VCSProviderID string `schema:"vcs_provider_id,required"`
	}

	// QualifiedName is the workspace's fully qualified name including the
	// name of its organization
	QualifiedName struct {
		Organization string
		Name         string
	}
)

func NewWorkspace(opts CreateOptions) (*Workspace, error) {
	// required options
	if opts.Name == nil {
		return nil, internal.ErrRequiredName
	}
	if opts.Organization == nil {
		return nil, internal.ErrRequiredOrg
	}

	ws := Workspace{
		ID:                  internal.NewID("ws"),
		CreatedAt:           internal.CurrentTimestamp(),
		UpdatedAt:           internal.CurrentTimestamp(),
		AllowDestroyPlan:    DefaultAllowDestroyPlan,
		ExecutionMode:       RemoteExecutionMode,
		FileTriggersEnabled: DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    DefaultTerraformVersion,
		SpeculativeEnabled:  true,
		Organization:        *opts.Organization,
	}
	if err := ws.setName(*opts.Name); err != nil {
		return nil, err
	}

	if opts.ExecutionMode != nil {
		if err := ws.setExecutionMode(*opts.ExecutionMode); err != nil {
			return nil, err
		}
	}
	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Branch != nil {
		ws.Branch = *opts.Branch
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.SourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.SourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		if err := ws.setTerraformVersion(*opts.TerraformVersion); err != nil {
			return nil, err
		}
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}
	return &ws, nil
}

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

func (ws *Workspace) String() string { return ws.Organization + "/" + ws.Name }

// ExecutionModes returns a list of possible execution modes
func (ws *Workspace) ExecutionModes() []string {
	return []string{"local", "remote", "agent"}
}

// QualifiedName returns the workspace's qualified name including the name of
// its organization
func (ws *Workspace) QualifiedName() QualifiedName {
	return QualifiedName{
		Organization: ws.Organization,
		Name:         ws.Name,
	}
}

// LogValue implements slog.LogValuer.
func (ws *Workspace) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", ws.ID),
		slog.String("organization", ws.Organization),
		slog.String("name", ws.Name),
	)
}

// Update updates the workspace with the given options.
func (ws *Workspace) Update(opts UpdateOptions) error {
	var updated bool

	if opts.Name != nil {
		if err := ws.setName(*opts.Name); err != nil {
			return err
		}
		updated = true
	}
	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
		updated = true
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
		updated = true
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
		updated = true
	}
	if opts.ExecutionMode != nil {
		if err := ws.setExecutionMode(*opts.ExecutionMode); err != nil {
			return err
		}
		updated = true
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
		updated = true
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.ExecutionMode = "remote"
		} else {
			ws.ExecutionMode = "local"
		}
		updated = true
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
		updated = true
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
		updated = true
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
		updated = true
	}
	if opts.TerraformVersion != nil {
		if err := ws.setTerraformVersion(*opts.TerraformVersion); err != nil {
			return err
		}
		updated = true
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
		updated = true
	}
	if updated {
		ws.UpdatedAt = internal.CurrentTimestamp()
	}

	return nil
}

func (ws *Workspace) setName(name string) error {
	if !internal.ReStringID.MatchString(name) {
		return internal.ErrInvalidName
	}
	ws.Name = name
	return nil
}

func (ws *Workspace) setExecutionMode(m ExecutionMode) error {
	if m != RemoteExecutionMode && m != LocalExecutionMode && m != AgentExecutionMode {
		return errors.New("invalid execution mode")
	}
	ws.ExecutionMode = m
	return nil
}

func (ws *Workspace) setTerraformVersion(v string) error {
	if !internal.ValidSemanticVersion(v) {
		return internal.ErrInvalidTerraformVersion
	}
	if result := semver.Compare(v, MinTerraformVersion); result < 0 {
		return internal.ErrUnsupportedTerraformVersion
	}
	ws.TerraformVersion = v
	return nil
}
