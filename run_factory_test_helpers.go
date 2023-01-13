package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRunCreateOptions is for testing purposes only.
type TestRunCreateOptions struct {
	ID                *string // override ID of run
	Speculative       bool
	ExecutionMode     *ExecutionMode
	Status            RunStatus
	AutoApply         bool
	Repo              *WorkspaceRepo
	IngressAttributes *IngressAttributes
}

// NewTestRun creates a new run. Expressly for testing purposes.
func NewTestRun(t *testing.T, opts TestRunCreateOptions) *Run {
	org, err := NewOrganization(OrganizationCreateOptions{Name: String("test-org")})
	require.NoError(t, err)

	ws, err := NewWorkspace(org, WorkspaceCreateOptions{
		Name:      "test-ws",
		AutoApply: Bool(opts.AutoApply),
		Repo:      opts.Repo,
	})
	require.NoError(t, err)

	cv, err := NewConfigurationVersion(ws.ID(), ConfigurationVersionCreateOptions{
		IngressAttributes: opts.IngressAttributes,
		Speculative:       Bool(opts.Speculative),
	})
	require.NoError(t, err)

	run := NewRun(cv, ws, RunCreateOptions{})
	if opts.Status != RunStatus("") {
		run.updateStatus(opts.Status)
	}
	if opts.ID != nil {
		run.id = *opts.ID
	}
	if opts.ExecutionMode != nil {
		run.executionMode = *opts.ExecutionMode
	}
	return run
}

type fakeRunFactoryWorkspaceService struct {
	ws *Workspace
	WorkspaceService
}

func (f *fakeRunFactoryWorkspaceService) GetWorkspace(context.Context, WorkspaceSpec) (*Workspace, error) {
	return f.ws, nil
}

type fakeRunFactoryConfigurationVersionService struct {
	cv *ConfigurationVersion
	ConfigurationVersionService
}

func (f *fakeRunFactoryConfigurationVersionService) GetConfigurationVersion(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeRunFactoryConfigurationVersionService) GetLatestConfigurationVersion(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}
