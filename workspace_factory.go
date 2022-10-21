package otf

import (
	"context"
)

type WorkspaceFactory struct {
	OrganizationService OrganizationService
}

func (f *WorkspaceFactory) NewWorkspace(ctx context.Context, opts WorkspaceCreateOptions) (*Workspace, error) {
	org, err := f.OrganizationService.GetOrganization(ctx, opts.OrganizationName)
	if err != nil {
		return nil, err
	}
	return NewWorkspace(org, opts)
}

func NewWorkspace(organization *Organization, opts WorkspaceCreateOptions) (*Workspace, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}
	ws := Workspace{
		id:                  NewID("ws"),
		name:                opts.Name,
		createdAt:           CurrentTimestamp(),
		updatedAt:           CurrentTimestamp(),
		allowDestroyPlan:    DefaultAllowDestroyPlan,
		executionMode:       RemoteExecutionMode,
		fileTriggersEnabled: DefaultFileTriggersEnabled,
		globalRemoteState:   true, // Only global remote state is supported
		terraformVersion:    DefaultTerraformVersion,
		speculativeEnabled:  true,
		lock:                &Unlocked{},
		organization:        organization,
	}
	// TODO: ExecutionMode and Operations are mututally exclusive options, this
	// should be enforced.
	if opts.ExecutionMode != nil {
		ws.executionMode = *opts.ExecutionMode
	}
	if opts.AllowDestroyPlan != nil {
		ws.allowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.autoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.fileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.QueueAllRuns != nil {
		ws.queueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.sourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.sourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.speculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.structuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.terraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.triggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.workingDirectory = *opts.WorkingDirectory
	}
	if opts.LatestRunID != nil {
		ws.latestRunID = opts.LatestRunID
	}
	return &ws, nil
}

// WorkspaceCreateOptions represents the options for creating a new workspace.
type WorkspaceCreateOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Description                *string
	ExecutionMode              *ExecutionMode
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	MigrationEnvironment       *string
	Name                       string
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	SourceName                 *string
	SourceURL                  *string
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string
	TriggerPrefixes            []string
	WorkingDirectory           *string
	OrganizationName           string `schema:"organization_name"`

	// Options for testing purposes only
	LatestRunID *string
}

func (o WorkspaceCreateOptions) Valid() error {
	if !ValidStringID(&o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}
	return nil
}
