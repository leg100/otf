package otf

import (
	"github.com/google/uuid"
	"github.com/mitchellh/copystructure"
)

func NewWorkspace(opts WorkspaceCreateOptions, org *Organization) *Workspace {
	ws := Workspace{
		ID:                  NewID("ws"),
		Name:                *opts.Name,
		AllowDestroyPlan:    DefaultAllowDestroyPlan,
		ExecutionMode:       DefaultExecutionMode,
		FileTriggersEnabled: DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    DefaultTerraformVersion,
		SpeculativeEnabled:  true,
		Organization:        org,
	}

	// TODO: ExecutionMode and Operations are mututally exclusive options, this
	// should be enforced.
	if opts.ExecutionMode != nil {
		ws.ExecutionMode = *opts.ExecutionMode
	}
	// Operations is deprecated in favour of ExecutionMode.
	if opts.Operations != nil {
		if *opts.Operations {
			ws.ExecutionMode = "remote"
		} else {
			ws.ExecutionMode = "local"
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
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	return &ws
}

func NewTestWorkspace(org *Organization) *Workspace {
	ws := Workspace{
		ID:           NewID("ws"),
		Name:         uuid.NewString(),
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
