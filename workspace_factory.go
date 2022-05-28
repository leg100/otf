package otf

import (
	"github.com/google/uuid"
)

type WorkspaceFactory struct {
	OrganizationService OrganizationService
}

func (f *WorkspaceFactory) NewWorkspace(opts WorkspaceCreateOptions) *Workspace {
	ws := Workspace{
		id:                  NewID("ws"),
		name:                opts.Name,
		createdAt:           CurrentTimestamp(),
		updatedAt:           CurrentTimestamp(),
		allowDestroyPlan:    DefaultAllowDestroyPlan,
		executionMode:       DefaultExecutionMode,
		fileTriggersEnabled: DefaultFileTriggersEnabled,
		globalRemoteState:   true, // Only global remote state is supported
		terraformVersion:    DefaultTerraformVersion,
		speculativeEnabled:  true,
		Organization:        &Organization{id: org.ID()},
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

	return &ws
}

func NewTestWorkspace(org *Organization) *Workspace {
	ws := Workspace{
		id:           NewID("ws"),
		name:         uuid.NewString(),
		Organization: &Organization{id: org.ID()},
		createdAt:    CurrentTimestamp(),
		updatedAt:    CurrentTimestamp(),
	}
	return &ws
}
