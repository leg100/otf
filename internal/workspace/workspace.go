// Package workspace provides access to terraform workspaces
package workspace

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/gobwas/glob"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/semver"
	"golang.org/x/exp/slog"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true

	MinTerraformVersion     = "1.2.0"
	DefaultTerraformVersion = "1.5.2"
)

var ErrNoVCSConnection = errors.New("workspace is not connected to a vcs repo")

type (
	// Workspace is a terraform workspace.
	Workspace struct {
		ID                         string        `json:"id"`
		CreatedAt                  time.Time     `json:"created_at"`
		UpdatedAt                  time.Time     `json:"updated_at"`
		AllowDestroyPlan           bool          `json:"allow_destroy_plan"`
		AutoApply                  bool          `json:"auto_apply"`
		CanQueueDestroyPlan        bool          `json:"can_queue_destroy_plan"`
		Description                string        `json:"description"`
		Environment                string        `json:"environment"`
		ExecutionMode              ExecutionMode `json:"execution_mode"`
		GlobalRemoteState          bool          `json:"global_remote_state"`
		MigrationEnvironment       string        `json:"migration_environment"`
		Name                       string        `json:"name"`
		QueueAllRuns               bool          `json:"queue_all_runs"`
		SpeculativeEnabled         bool          `json:"speculative_enabled"`
		StructuredRunOutputEnabled bool          `json:"structured_run_output_enabled"`
		SourceName                 string        `json:"source_name"`
		SourceURL                  string        `json:"source_url"`
		TerraformVersion           string        `json:"terraform_version"`
		WorkingDirectory           string        `json:"working_directory"`
		Organization               string        `json:"organization"`
		LatestRun                  *LatestRun    `json:"latest_run"`
		Tags                       []string      `json:"tags"`
		Lock                       *Lock         `json:"lock"`

		// VCS connection fields, which ought to belong in Connection but are
		// included at the root of the workspace because the go-tfe integration
		// tests set these fields without setting the connection!
		FileTriggersEnabled bool     `json:"file_triggers_enabled"`
		TriggerPrefixes     []string `json:"trigger_prefixes"`
		TriggerPatterns     []string `json:"trigger_patterns"`

		Connection *Connection
	}

	Connection struct {
		Branch    *string `json:"branch"`
		TagsRegex *string `json:"tags_regex"`

		VCSProviderID string
		Repo          string
	}

	ConnectOptions struct {
		Branch        *string
		RepoPath      *string
		VCSProviderID *string
		TagsRegex     *string
	}

	// LatestRun is a summary of the latest run for a workspace
	LatestRun struct {
		ID     string
		Status internal.RunStatus
	}

	ExecutionMode string

	// CreateOptions represents the options for creating a new workspace.
	CreateOptions struct {
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Description                *string
		ExecutionMode              *ExecutionMode
		FileTriggersEnabled        *bool
		GlobalRemoteState          *bool
		MigrationEnvironment       *string
		Name                       *string
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		SourceName                 *string
		SourceURL                  *string
		StructuredRunOutputEnabled *bool
		Tags                       []TagSpec
		TerraformVersion           *string
		TriggerPrefixes            []string
		TriggerPatterns            []string
		WorkingDirectory           *string
		Organization               *string

		*ConnectOptions
	}

	UpdateOptions struct {
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Name                       *string
		Description                *string
		ExecutionMode              *ExecutionMode
		FileTriggersEnabled        *bool
		GlobalRemoteState          *bool
		Operations                 *bool
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		StructuredRunOutputEnabled *bool
		TerraformVersion           *string
		TriggerPrefixes            []string
		TriggerPatterns            []string
		WorkingDirectory           *string

		// Disconnect workspace from repo. It is invalid to specify true for an
		// already disconnected workspace.
		Disconnect bool

		// Specifying ConnectOptions either connects a currently
		// disconnected workspace, or modifies a connection if already
		// connected.
		*ConnectOptions
	}

	// ListOptions are options for paginating and filtering a list of
	// Workspaces
	ListOptions struct {
		Search       string
		Tags         []string
		Organization *string

		resource.PageOptions
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
	if opts.TriggerPrefixes != nil && opts.TriggerPatterns != nil {
		return nil, errors.New("cannot specify both trigger prefixes and trigger patterns")
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.TriggerPatterns != nil {
		if err := ws.setTriggerPatterns(opts.TriggerPatterns); err != nil {
			return nil, err
		}
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}
	if opts.ConnectOptions != nil {
		if err := ws.addConnection(opts.ConnectOptions); err != nil {
			return nil, err
		}
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

// LogValue implements slog.LogValuer.
func (ws *Workspace) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", ws.ID),
		slog.String("organization", ws.Organization),
		slog.String("name", ws.Name),
	)
}

// Update updates the workspace with the given options. A boolean is returned to
// indicate whether the workspace is to be connected to a repo (true),
// disconnected from a repo (false), or neither (nil).
func (ws *Workspace) Update(opts UpdateOptions) (*bool, error) {
	var updated bool

	if opts.Name != nil {
		if err := ws.setName(*opts.Name); err != nil {
			return nil, err
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
			return nil, err
		}
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
			return nil, err
		}
		updated = true
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
		updated = true
	}

	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
		updated = true
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	if opts.TriggerPatterns != nil {
		if err := ws.setTriggerPatterns(opts.TriggerPatterns); err != nil {
			return nil, err
		}
		updated = true
	}

	// determine whether to connect or disconnect workspace
	var connect *bool
	if opts.Disconnect {
		if ws.Connection == nil {
			return nil, errors.New("cannot disconnect an already disconnected workspace")
		}
		// workspace is to be disconnected
		connect = internal.Bool(false)
		updated = true
	} else if opts.ConnectOptions != nil && ws.Connection == nil {
		// workspace is to be connected
		if err := ws.addConnection(opts.ConnectOptions); err != nil {
			return nil, err
		}
		connect = internal.Bool(true)
		updated = true
	}

	if updated {
		ws.UpdatedAt = internal.CurrentTimestamp()
	}

	return connect, nil
}

func (ws *Workspace) addConnection(opts *ConnectOptions) error {
	// must specify both repo and vcs provider ID
	if opts.RepoPath == nil {
		return &internal.MissingParameterError{Parameter: "repo_path"}
	}
	if opts.VCSProviderID == nil {
		return &internal.MissingParameterError{Parameter: "vcs_provider_id"}
	}
	if opts.TagsRegex != nil {
		if _, err := regexp.Compile(*opts.TagsRegex); err != nil {
			return fmt.Errorf("invalid VCS tags regex: %w", err)
		}
	}
	ws.Connection = &Connection{
		Repo:          *opts.RepoPath,
		VCSProviderID: *opts.VCSProviderID,
		Branch:        opts.Branch,
		TagsRegex:     opts.TagsRegex,
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
	if !semver.IsValid(v) {
		return internal.ErrInvalidTerraformVersion
	}
	if result := semver.Compare(v, MinTerraformVersion); result < 0 {
		return internal.ErrUnsupportedTerraformVersion
	}
	ws.TerraformVersion = v
	return nil
}

func (ws *Workspace) setTriggerPatterns(patterns []string) error {
	for _, patt := range patterns {
		if _, err := glob.Compile(patt); err != nil {
			return fmt.Errorf("invalid trigger pattern: %w", err)
		}
	}
	ws.TriggerPatterns = patterns
	return nil
}
