// Package workspace is responsible for terraform workspaces
package workspace

import (
	"errors"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/semver"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	id                         string
	createdAt                  time.Time
	updatedAt                  time.Time
	allowDestroyPlan           bool
	autoApply                  bool
	canQueueDestroyPlan        bool
	description                string
	environment                string
	executionMode              otf.ExecutionMode
	fileTriggersEnabled        bool
	globalRemoteState          bool
	migrationEnvironment       string
	name                       string
	queueAllRuns               bool
	speculativeEnabled         bool
	structuredRunOutputEnabled bool
	sourceName                 string
	sourceURL                  string
	terraformVersion           string
	triggerPrefixes            []string
	workingDirectory           string
	organization               string
	latestRunID                *string
	repo                       *WorkspaceRepo
	permissions                []otf.WorkspacePermission

	Lock
}

func (ws *Workspace) ID() string                       { return ws.id }
func (ws *Workspace) CreatedAt() time.Time             { return ws.createdAt }
func (ws *Workspace) UpdatedAt() time.Time             { return ws.updatedAt }
func (ws *Workspace) String() string                   { return ws.organization + "/" + ws.name }
func (ws *Workspace) Name() string                     { return ws.name }
func (ws *Workspace) AllowDestroyPlan() bool           { return ws.allowDestroyPlan }
func (ws *Workspace) AutoApply() bool                  { return ws.autoApply }
func (ws *Workspace) CanQueueDestroyPlan() bool        { return ws.canQueueDestroyPlan }
func (ws *Workspace) Environment() string              { return ws.environment }
func (ws *Workspace) Description() string              { return ws.description }
func (ws *Workspace) ExecutionMode() otf.ExecutionMode { return ws.executionMode }
func (ws *Workspace) FileTriggersEnabled() bool        { return ws.fileTriggersEnabled }
func (ws *Workspace) GlobalRemoteState() bool          { return ws.globalRemoteState }
func (ws *Workspace) MigrationEnvironment() string     { return ws.migrationEnvironment }
func (ws *Workspace) QueueAllRuns() bool               { return ws.queueAllRuns }
func (ws *Workspace) SourceName() string               { return ws.sourceName }
func (ws *Workspace) SourceURL() string                { return ws.sourceURL }
func (ws *Workspace) SpeculativeEnabled() bool         { return ws.speculativeEnabled }
func (ws *Workspace) StructuredRunOutputEnabled() bool { return ws.structuredRunOutputEnabled }
func (ws *Workspace) TerraformVersion() string         { return ws.terraformVersion }
func (ws *Workspace) TriggerPrefixes() []string        { return ws.triggerPrefixes }
func (ws *Workspace) WorkingDirectory() string         { return ws.workingDirectory }
func (ws *Workspace) Organization() string             { return ws.organization }
func (ws *Workspace) LatestRunID() *string             { return ws.latestRunID }
func (ws *Workspace) Repo() *WorkspaceRepo             { return ws.repo }

// WorkspaceID returns the workspace's ID. Implemented in order to satisfy
// WorkspaceResource.
func (ws *Workspace) WorkspaceID() string { return ws.id }

func (ws *Workspace) SetLatestRun(runID string) {
	ws.latestRunID = otf.String(runID)
}

// ExecutionModes returns a list of possible execution modes
func (ws *Workspace) ExecutionModes() []string {
	return []string{"local", "remote", "agent"}
}

// QualifiedName returns the workspace's qualified name including the name of
// its organization
func (ws *Workspace) QualifiedName() WorkspaceQualifiedName {
	return WorkspaceQualifiedName{
		Organization: ws.Organization(),
		Name:         ws.Name(),
	}
}

func (ws *Workspace) MarshalLog() any {
	log := struct {
		Name         string `json:"name"`
		Organization string `json:"organization"`
	}{
		Name:         ws.name,
		Organization: ws.organization,
	}
	return log
}

// Update updates the workspace with the given options.
func (ws *Workspace) Update(opts UpdateWorkspaceOptions) error {
	var updated bool

	if opts.Name != nil {
		if err := ws.setName(*opts.Name); err != nil {
			return err
		}
		updated = true
	}
	if opts.AllowDestroyPlan != nil {
		ws.allowDestroyPlan = *opts.AllowDestroyPlan
		updated = true
	}
	if opts.AutoApply != nil {
		ws.autoApply = *opts.AutoApply
		updated = true
	}
	if opts.Description != nil {
		ws.description = *opts.Description
		updated = true
	}
	if opts.ExecutionMode != nil {
		if err := ws.setExecutionMode(*opts.ExecutionMode); err != nil {
			return err
		}
		updated = true
	}
	if opts.FileTriggersEnabled != nil {
		ws.fileTriggersEnabled = *opts.FileTriggersEnabled
		updated = true
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.executionMode = "remote"
		} else {
			ws.executionMode = "local"
		}
		updated = true
	}
	if opts.QueueAllRuns != nil {
		ws.queueAllRuns = *opts.QueueAllRuns
		updated = true
	}
	if opts.SpeculativeEnabled != nil {
		ws.speculativeEnabled = *opts.SpeculativeEnabled
		updated = true
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.structuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
		updated = true
	}
	if opts.TerraformVersion != nil {
		if err := ws.setTerraformVersion(*opts.TerraformVersion); err != nil {
			return err
		}
		updated = true
	}
	if opts.TriggerPrefixes != nil {
		ws.triggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	if opts.WorkingDirectory != nil {
		ws.workingDirectory = *opts.WorkingDirectory
		updated = true
	}
	if updated {
		ws.updatedAt = otf.CurrentTimestamp()
	}

	return nil
}

func (ws *Workspace) setName(name string) error {
	if !otf.ReStringID.MatchString(name) {
		return otf.ErrInvalidName
	}
	ws.name = name
	return nil
}

func (ws *Workspace) setExecutionMode(m otf.ExecutionMode) error {
	if m != otf.RemoteExecutionMode && m != otf.LocalExecutionMode && m != otf.AgentExecutionMode {
		return errors.New("invalid execution mode")
	}
	ws.executionMode = m
	return nil
}

func (ws *Workspace) setTerraformVersion(v string) error {
	if !otf.ValidSemanticVersion(v) {
		return otf.ErrInvalidTerraformVersion
	}
	if result := semver.Compare(v, otf.MinTerraformVersion); result < 0 {
		return otf.ErrUnsupportedTerraformVersion
	}
	ws.terraformVersion = v
	return nil
}

// WorkspaceQualifiedName is the workspace's fully qualified name including the
// name of its organization
type WorkspaceQualifiedName struct {
	Organization string
	Name         string
}

// UpdateWorkspaceOptions represents the options for updating a workspace.
type UpdateWorkspaceOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Name                       *string
	Description                *string
	ExecutionMode              *otf.ExecutionMode `schema:"execution_mode"`
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
