package policy

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type EnforcementLevel string
type PolicySetSource string

const (
	AdvisoryEnforcement  EnforcementLevel = "advisory"
	MandatoryEnforcement EnforcementLevel = "mandatory"

	ManualPolicySetSource PolicySetSource = "manual"
	VCSPolicySetSource    PolicySetSource = "vcs"
)

type PolicySet struct {
	ID             resource.TfeID    `db:"policy_set_id"`
	CreatedAt      time.Time         `db:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at"`
	Organization   organization.Name `db:"organization_name"`
	Name           string            `db:"name"`
	Description    string            `db:"description"`
	Source         PolicySetSource   `db:"source"`
	VCSProviderID  *resource.TfeID   `db:"vcs_provider_id"`
	VCSRepo        *vcs.Repo         `db:"vcs_repo"`
	VCSRef         string            `db:"vcs_ref"`
	VCSPath        string            `db:"vcs_path"`
	VCSPolicyPaths []string          `db:"vcs_policy_paths"`
	LastSyncedAt   *time.Time        `db:"last_synced_at"`
}

type Policy struct {
	ID               resource.TfeID    `db:"policy_id"`
	PolicySetID      resource.TfeID    `db:"policy_set_id"`
	CreatedAt        time.Time         `db:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at"`
	Name             string            `db:"name"`
	Description      string            `db:"description"`
	EnforcementLevel EnforcementLevel  `db:"enforcement_level"`
	Source           string            `db:"source"`
	Path             string            `db:"path"`
	Organization     organization.Name `db:"organization_name"`
	PolicySetName    string            `db:"policy_set_name"`
}

type PolicyModule struct {
	ID          resource.TfeID `db:"policy_module_id"`
	PolicySetID resource.TfeID `db:"policy_set_id"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	Name        string         `db:"name"`
	Path        string         `db:"path"`
	Source      string         `db:"source"`
}

type PolicyCheck struct {
	ID               resource.TfeID    `db:"policy_check_id"`
	RunID            resource.TfeID    `db:"run_id"`
	WorkspaceID      resource.TfeID    `db:"workspace_id"`
	PolicySetID      resource.TfeID    `db:"policy_set_id"`
	PolicyID         resource.TfeID    `db:"policy_id"`
	Organization     organization.Name `db:"organization_name"`
	PolicyName       string            `db:"policy_name"`
	PolicySetName    string            `db:"policy_set_name"`
	EnforcementLevel EnforcementLevel  `db:"enforcement_level"`
	Passed           bool              `db:"passed"`
	Output           string            `db:"output"`
	CreatedAt        time.Time         `db:"created_at"`
}

type Variable struct {
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Category  string `json:"category"`
	Sensitive bool   `json:"sensitive"`
	HCL       bool   `json:"hcl"`
}

type RunVariableInfo struct {
	Category  string `json:"category"`
	Sensitive bool   `json:"sensitive"`
}

type TagBinding struct {
	Key       string  `json:"key"`
	Value     *string `json:"value"`
	Inherited bool    `json:"inherited"`
}

type RunInfo struct {
	ID                     string                     `json:"id"`
	CreatedAt              string                     `json:"created_at"`
	CreatedBy              string                     `json:"created_by,omitempty"`
	Message                string                     `json:"message,omitempty"`
	CommitSHA              string                     `json:"commit_sha,omitempty"`
	IsDestroy              bool                       `json:"is_destroy"`
	Refresh                bool                       `json:"refresh"`
	RefreshOnly            bool                       `json:"refresh_only"`
	ReplaceAddrs           []string                   `json:"replace_addrs,omitempty"`
	Speculative            bool                       `json:"speculative"`
	TargetAddrs            []string                   `json:"target_addrs,omitempty"`
	ConfigurationVersionID resource.TfeID             `json:"-"`
	Variables              map[string]RunVariableInfo `json:"variables,omitempty"`
	Project                struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Organization struct {
		Name string `json:"name"`
	} `json:"organization"`
	Workspace struct {
		ID               string         `json:"id"`
		Name             string         `json:"name"`
		CreatedAt        string         `json:"created_at"`
		Description      string         `json:"description,omitempty"`
		ExecutionMode    string         `json:"execution_mode"`
		AutoApply        bool           `json:"auto_apply"`
		WorkingDirectory string         `json:"working_directory,omitempty"`
		Tags             []string       `json:"tags,omitempty"`
		TagBindings      []TagBinding   `json:"tag_bindings,omitempty"`
		VCSRepo          map[string]any `json:"vcs_repo,omitempty"`
	} `json:"workspace"`
	CostEstimate *struct {
		PriorMonthlyCost    string `json:"prior_monthly_cost"`
		ProposedMonthlyCost string `json:"proposed_monthly_cost"`
		DeltaMonthlyCost    string `json:"delta_monthly_cost"`
	} `json:"cost_estimate,omitempty"`
}

type MockBundle struct {
	Files map[string][]byte
}

type EvaluationResult struct {
	Evaluated           bool
	Checks              []*PolicyCheck
	HasMandatoryFailure bool
	HasAdvisoryFailure  bool
}

type CreatePolicySetOptions struct {
	Name           *string
	Description    *string
	Source         PolicySetSource
	VCSProviderID  *resource.TfeID
	VCSRepo        *vcs.Repo
	VCSRef         *string
	VCSPath        *string
	VCSPolicyPaths []string
	LastSyncedAt   *time.Time
}

type UpdatePolicySetOptions struct {
	Name        *string
	Description *string
}

type CreatePolicyOptions struct {
	Name             *string
	Description      *string
	EnforcementLevel *EnforcementLevel
	Source           *string
	Path             *string
}

type UpdatePolicyOptions struct {
	Name             *string
	Description      *string
	EnforcementLevel *EnforcementLevel
	Source           *string
}

type workspaceClient interface {
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
}

type stateClient interface {
	DownloadCurrentState(context.Context, resource.TfeID) ([]byte, error)
}

type planClient interface {
	GetRunPlanJSON(context.Context, resource.TfeID) ([]byte, error)
}

type variableClient interface {
	ListRunVariables(context.Context, resource.TfeID) ([]Variable, error)
}

type runClient interface {
	GetRunInfo(context.Context, resource.TfeID) (*RunInfo, error)
}

type configClient interface {
	DownloadConfig(context.Context, resource.TfeID) ([]byte, error)
	GetConfigVersion(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error)
}

type vcsProviderClient interface {
	GetVCSProvider(context.Context, resource.TfeID) (*vcs.Provider, error)
}

type repoHookClient interface {
	CreateRepohook(context.Context, repohooks.CreateRepohookOptions) (uuid.UUID, error)
	DeleteUnreferencedRepohooks(context.Context) error
}

type ImportablePolicy struct {
	Name             string
	Path             string
	Source           string
	EnforcementLevel EnforcementLevel
}

type ImportableModule struct {
	Name   string
	Path   string
	Source string
}

type SyncResult struct {
	Set      *PolicySet
	Imported []*Policy
}

type CreateVCSPolicySetOptions struct {
	Name                *string
	Description         *string
	VCSProviderID       *resource.TfeID
	VCSRepo             *vcs.Repo
	VCSRef              *string
	VCSPath             *string
	SelectedPolicyPaths []string
}

type Evaluator interface {
	Evaluate(context.Context, *MockBundle, []*Policy, []*PolicyModule) ([]*PolicyCheck, error)
}

func newPolicySet(org organization.Name, opts CreatePolicySetOptions) (*PolicySet, error) {
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	now := internal.CurrentTimestamp(nil)
	set := &PolicySet{
		ID:             resource.NewTfeID(resource.PolicySetKind),
		CreatedAt:      now,
		UpdatedAt:      now,
		Organization:   org,
		Name:           *opts.Name,
		Source:         ManualPolicySetSource,
		VCSPolicyPaths: []string{},
	}
	if opts.Description != nil {
		set.Description = *opts.Description
	}
	if opts.Source != "" {
		set.Source = opts.Source
	}
	if opts.VCSProviderID != nil {
		set.VCSProviderID = opts.VCSProviderID
	}
	if opts.VCSRepo != nil {
		set.VCSRepo = opts.VCSRepo
	}
	if opts.VCSRef != nil {
		set.VCSRef = *opts.VCSRef
	}
	if opts.VCSPath != nil {
		set.VCSPath = *opts.VCSPath
	}
	if opts.VCSPolicyPaths != nil {
		set.VCSPolicyPaths = append([]string(nil), opts.VCSPolicyPaths...)
	}
	if opts.LastSyncedAt != nil {
		set.LastSyncedAt = opts.LastSyncedAt
	}
	return set, nil
}

func newPolicy(set *PolicySet, opts CreatePolicyOptions) (*Policy, error) {
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	if opts.EnforcementLevel == nil {
		return nil, fmt.Errorf("missing enforcement level")
	}
	if *opts.EnforcementLevel != AdvisoryEnforcement && *opts.EnforcementLevel != MandatoryEnforcement {
		return nil, fmt.Errorf("invalid enforcement level: %s", *opts.EnforcementLevel)
	}
	if opts.Source == nil || *opts.Source == "" {
		return nil, fmt.Errorf("missing source")
	}
	now := internal.CurrentTimestamp(nil)
	policy := &Policy{
		ID:               resource.NewTfeID(resource.PolicyKind),
		PolicySetID:      set.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
		Name:             *opts.Name,
		EnforcementLevel: *opts.EnforcementLevel,
		Source:           *opts.Source,
		Path:             "",
		Organization:     set.Organization,
		PolicySetName:    set.Name,
	}
	if opts.Description != nil {
		policy.Description = *opts.Description
	}
	if opts.Path != nil {
		policy.Path = *opts.Path
	}
	return policy, nil
}

func (b *MockBundle) Zip() ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, data := range b.Files {
		f, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write(data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalJSON(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
