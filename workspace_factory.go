package otf

import (
	"github.com/google/uuid"
	"github.com/mitchellh/copystructure"
)

func NewWorkspace(opts WorkspaceCreateOptions, org *Organization) *Workspace {
	ws := Workspace{
		ID:                  NewID("ws"),
		name:                *opts.Name,
		allowDestroyPlan:    DefaultAllowDestroyPlan,
		executionMode:       DefaultExecutionMode,
		fileTriggersEnabled: DefaultFileTriggersEnabled,
		globalRemoteState:   true, // Only global remote state is supported
		terraformVersion:    DefaultTerraformVersion,
		speculativeEnabled:  true,
		Organization:        org,
	}

	// TODO: ExecutionMode and Operations are mututally exclusive options, this
	// should be enforced.
	if opts.ExecutionMode != nil {
		ws.executionMode = *opts.ExecutionMode
	}
	// Operations is deprecated in favour of ExecutionMode.
	if opts.Operations != nil {
		if *opts.Operations {
			ws.executionMode = "remote"
		} else {
			ws.executionMode = "local"
		}
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
		ID:           NewID("ws"),
		name:         uuid.NewString(),
		Organization: org,
	}
	return &ws
}

func NewShallowNestedWorkspace(ws *Workspace) *Workspace {
	cp, _ := copystructure.Copy(ws)
	shallowWorkspace := cp.(*Workspace)
	shallowWorkspace.Organization = &Organization{ID: ws.Organization.ID}
	return shallowWorkspace
}
