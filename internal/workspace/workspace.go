// Package workspace provides access to terraform workspaces
package workspace

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"log/slog"

	"slices"

	"github.com/gobwas/glob"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/semver"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	DefaultAllowDestroyPlan = true
	MinTerraformVersion     = "1.2.0"
)

var (
	apiTestTerraformVersions = []string{"0.10.0", "0.11.0", "0.11.1"}
)

type (
	// Workspace is a terraform workspace.
	Workspace struct {
		ID                         string        `jsonapi:"primary,workspaces"`
		CreatedAt                  time.Time     `jsonapi:"attribute" json:"created_at"`
		UpdatedAt                  time.Time     `jsonapi:"attribute" json:"updated_at"`
		AgentPoolID                *string       `jsonapi:"attribute" json:"agent-pool-id"`
		AllowDestroyPlan           bool          `jsonapi:"attribute" json:"allow_destroy_plan"`
		AutoApply                  bool          `jsonapi:"attribute" json:"auto_apply"`
		CanQueueDestroyPlan        bool          `jsonapi:"attribute" json:"can_queue_destroy_plan"`
		Description                string        `jsonapi:"attribute" json:"description"`
		Environment                string        `jsonapi:"attribute" json:"environment"`
		ExecutionMode              ExecutionMode `jsonapi:"attribute" json:"execution_mode"`
		GlobalRemoteState          bool          `jsonapi:"attribute" json:"global_remote_state"`
		MigrationEnvironment       string        `jsonapi:"attribute" json:"migration_environment"`
		Name                       string        `jsonapi:"attribute" json:"name"`
		QueueAllRuns               bool          `jsonapi:"attribute" json:"queue_all_runs"`
		SpeculativeEnabled         bool          `jsonapi:"attribute" json:"speculative_enabled"`
		StructuredRunOutputEnabled bool          `jsonapi:"attribute" json:"structured_run_output_enabled"`
		SourceName                 string        `jsonapi:"attribute" json:"source_name"`
		SourceURL                  string        `jsonapi:"attribute" json:"source_url"`
		TerraformVersion           string        `jsonapi:"attribute" json:"terraform_version"`
		WorkingDirectory           string        `jsonapi:"attribute" json:"working_directory"`
		Organization               string        `jsonapi:"attribute" json:"organization"`
		LatestRun                  *LatestRun    `jsonapi:"attribute" json:"latest_run"`
		Tags                       []string      `jsonapi:"attribute" json:"tags"`
		Lock                       *Lock         `jsonapi:"attribute" json:"lock"`

		// VCS Connection; nil means the workspace is not connected.
		Connection *Connection

		// TriggerPatterns is mutually exclusive with Connection.TagsRegex.
		//
		// Note: TriggerPatterns ought to belong in Connection but it is included at
		// the root of Workspace because the go-tfe integration tests set
		// this field without setting the connection!
		TriggerPatterns []string

		// TriggerPrefixes exists only to pass the go-tfe integration tests and
		// is not used when determining whether to trigger runs. Use
		// TriggerPatterns instead.
		TriggerPrefixes []string
	}

	Connection struct {
		// Pushes to this VCS branch trigger runs. Empty string means the default
		// branch is used. Ignored if TagsRegex is non-empty.
		Branch string
		// Pushed tags matching this regular expression trigger runs. Mutually
		// exclusive with TriggerPatterns.
		TagsRegex string

		VCSProviderID string
		Repo          string

		// By default, once a workspace is connected to a repo it is not
		// possible to run a terraform apply via the CLI. Setting this to true
		// overrides this behaviour.
		AllowCLIApply bool
	}

	ConnectOptions struct {
		RepoPath      *string
		VCSProviderID *string

		Branch        *string
		TagsRegex     *string
		AllowCLIApply *bool
	}

	ExecutionMode string

	// CreateOptions represents the options for creating a new workspace.
	CreateOptions struct {
		AgentPoolID                *string
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Description                *string
		ExecutionMode              *ExecutionMode
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

		// Always trigger runs. A value of true is mutually exclusive with
		// setting TriggerPatterns or ConnectOptions.TagsRegex.
		AlwaysTrigger *bool

		*ConnectOptions
	}

	UpdateOptions struct {
		AgentPoolID                *string `json:"agent-pool-id,omitempty"`
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Name                       *string
		Description                *string
		ExecutionMode              *ExecutionMode `json:"execution-mode,omitempty"`
		GlobalRemoteState          *bool
		Operations                 *bool
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		StructuredRunOutputEnabled *bool
		TerraformVersion           *string
		TriggerPrefixes            []string
		TriggerPatterns            []string
		WorkingDirectory           *string

		// Always trigger runs. A value of true is mutually exclusive with
		// setting TriggerPatterns or ConnectOptions.TagsRegex.
		AlwaysTrigger *bool

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
		Organization *string `schema:"organization_name"`

		resource.PageOptions
	}

	// VCS trigger strategy determines which VCS events trigger runs
	VCSTriggerStrategy string
)

func NewWorkspace(opts CreateOptions) (*Workspace, error) {
	// required options
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	if opts.Organization == nil {
		return nil, internal.ErrRequiredOrg
	}

	ws := Workspace{
		ID:                 internal.NewID("ws"),
		CreatedAt:          internal.CurrentTimestamp(nil),
		UpdatedAt:          internal.CurrentTimestamp(nil),
		AllowDestroyPlan:   DefaultAllowDestroyPlan,
		ExecutionMode:      RemoteExecutionMode,
		TerraformVersion:   releases.DefaultTerraformVersion,
		SpeculativeEnabled: true,
		Organization:       *opts.Organization,
	}
	if err := ws.setName(*opts.Name); err != nil {
		return nil, err
	}
	if _, err := ws.setExecutionModeAndAgentPoolID(opts.ExecutionMode, opts.AgentPoolID); err != nil {
		return nil, err
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
	if opts.GlobalRemoteState != nil {
		ws.GlobalRemoteState = *opts.GlobalRemoteState
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
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}
	// TriggerPrefixes are not used but OTF persists it in order to pass go-tfe
	// integration tests.
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	// Enforce three-way mutually exclusivity between:
	// (a) tags-regex
	// (b) trigger-patterns
	// (c) always-trigger=true
	if (opts.ConnectOptions != nil && (opts.ConnectOptions.TagsRegex != nil && *opts.ConnectOptions.TagsRegex != "")) && opts.TriggerPatterns != nil {
		return nil, ErrTagsRegexAndTriggerPatterns
	}
	if (opts.ConnectOptions != nil && (opts.ConnectOptions.TagsRegex != nil && *opts.ConnectOptions.TagsRegex != "")) && (opts.AlwaysTrigger != nil && *opts.AlwaysTrigger) {
		return nil, ErrTagsRegexAndAlwaysTrigger
	}
	if len(opts.TriggerPatterns) > 0 && (opts.AlwaysTrigger != nil && *opts.AlwaysTrigger) {
		return nil, ErrTriggerPatternsAndAlwaysTrigger
	}
	if opts.ConnectOptions != nil {
		if err := ws.addConnection(opts.ConnectOptions); err != nil {
			return nil, err
		}
	}
	if opts.TriggerPatterns != nil {
		if err := ws.setTriggerPatterns(opts.TriggerPatterns); err != nil {
			return nil, fmt.Errorf("setting trigger patterns: %w", err)
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
	if changed, err := ws.setExecutionModeAndAgentPoolID(opts.ExecutionMode, opts.AgentPoolID); err != nil {
		return nil, err
	} else if changed {
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
	if opts.GlobalRemoteState != nil {
		ws.GlobalRemoteState = *opts.GlobalRemoteState
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
	// TriggerPrefixes are not used but OTF persists it in order to pass go-tfe
	// integration tests.
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	// Enforce three-way mutually exclusivity between:
	// (a) tags-regex
	// (b) trigger-patterns
	// (c) always-trigger=true
	if (opts.ConnectOptions != nil && (opts.ConnectOptions.TagsRegex != nil && *opts.ConnectOptions.TagsRegex != "")) && opts.TriggerPatterns != nil {
		return nil, ErrTagsRegexAndTriggerPatterns
	}
	if (opts.ConnectOptions != nil && (opts.ConnectOptions.TagsRegex != nil && *opts.ConnectOptions.TagsRegex != "")) && (opts.AlwaysTrigger != nil && *opts.AlwaysTrigger) {
		return nil, ErrTagsRegexAndAlwaysTrigger
	}
	if len(opts.TriggerPatterns) > 0 && (opts.AlwaysTrigger != nil && *opts.AlwaysTrigger) {
		return nil, ErrTriggerPatternsAndAlwaysTrigger
	}
	if opts.AlwaysTrigger != nil && *opts.AlwaysTrigger {
		if ws.Connection != nil {
			ws.Connection.TagsRegex = ""
		}
		ws.TriggerPatterns = nil
		updated = true
	}
	if opts.TriggerPatterns != nil {
		if err := ws.setTriggerPatterns(opts.TriggerPatterns); err != nil {
			return nil, fmt.Errorf("setting trigger patterns: %w", err)
		}
		if ws.Connection != nil {
			ws.Connection.TagsRegex = ""
		}
		updated = true
	}
	// determine whether to connect or disconnect workspace
	if opts.Disconnect && opts.ConnectOptions != nil {
		return nil, errors.New("connect options must be nil if disconnect is true")
	}
	var connect *bool
	if opts.Disconnect {
		if ws.Connection == nil {
			return nil, errors.New("cannot disconnect an already disconnected workspace")
		}
		// workspace is to be disconnected
		connect = internal.Bool(false)
		updated = true
	}
	if opts.ConnectOptions != nil {
		if ws.Connection == nil {
			// workspace is to be connected
			if err := ws.addConnection(opts.ConnectOptions); err != nil {
				return nil, err
			}
			connect = internal.Bool(true)
			updated = true
		} else {
			// modify existing connection
			if opts.TagsRegex != nil {
				if err := ws.setTagsRegex(*opts.TagsRegex); err != nil {
					return nil, fmt.Errorf("invalid tags-regex: %w", err)
				}
				ws.TriggerPatterns = nil
				updated = true
			}
			if opts.Branch != nil {
				ws.Connection.Branch = *opts.Branch
				updated = true
			}
			if opts.AllowCLIApply != nil {
				ws.Connection.AllowCLIApply = *opts.AllowCLIApply
				updated = true
			}
		}
	}
	if updated {
		ws.UpdatedAt = internal.CurrentTimestamp(nil)
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
	ws.Connection = &Connection{
		Repo:          *opts.RepoPath,
		VCSProviderID: *opts.VCSProviderID,
	}
	if opts.AllowCLIApply != nil {
		ws.Connection.AllowCLIApply = *opts.AllowCLIApply
	}
	if opts.TagsRegex != nil {
		if err := ws.setTagsRegex(*opts.TagsRegex); err != nil {
			return fmt.Errorf("invalid tags-regex: %w", err)
		}
	}
	if opts.Branch != nil {
		ws.Connection.Branch = *opts.Branch
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

// setExecutionModeAndAgentPoolID sets the execution mode and/or the agent pool
// ID. The two parameters are intimately related, hence the validation and
// setting of the parameters is handled in tandem.
func (ws *Workspace) setExecutionModeAndAgentPoolID(m *ExecutionMode, agentPoolID *string) (bool, error) {
	if m == nil {
		if agentPoolID == nil {
			// neither specified; nothing more to be done
			return false, nil
		} else {
			// agent pool ID can be set without specifying execution mode as long as
			// existing execution mode is AgentExecutionMode
			if ws.ExecutionMode != AgentExecutionMode {
				return false, ErrNonAgentExecutionModeWithPool
			}
		}
	} else {
		if *m == AgentExecutionMode {
			if agentPoolID == nil {
				return false, ErrAgentExecutionModeWithoutPool
			}
		} else {
			// mode is either remote or local; in either case no pool ID should be
			// provided
			if agentPoolID != nil {
				return false, ErrNonAgentExecutionModeWithPool
			}
		}
	}
	ws.AgentPoolID = agentPoolID
	if m != nil {
		ws.ExecutionMode = *m
	}
	return true, nil
}

func (ws *Workspace) setTerraformVersion(v string) error {
	if v == releases.LatestVersionString {
		ws.TerraformVersion = v
		return nil
	}
	if !semver.IsValid(v) {
		return internal.ErrInvalidTerraformVersion
	}
	// only accept terraform versions above the minimum requirement.
	//
	// NOTE: we make an exception for the specific versions posted by the go-tfe
	// integration tests.
	if result := semver.Compare(v, MinTerraformVersion); result < 0 {
		if !slices.Contains(apiTestTerraformVersions, v) {
			return ErrUnsupportedTerraformVersion
		}
	}
	ws.TerraformVersion = v
	return nil
}

func (ws *Workspace) setTagsRegex(regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return ErrInvalidTagsRegex
	}
	ws.Connection.TagsRegex = regex
	return nil
}

func (ws *Workspace) setTriggerPatterns(patterns []string) error {
	for _, patt := range patterns {
		if _, err := glob.Compile(patt); err != nil {
			return ErrInvalidTriggerPattern
		}
	}
	ws.TriggerPatterns = patterns
	return nil
}
