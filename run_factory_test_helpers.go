package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewTestRun creates a new run. Expressly for testing purposes.
func NewTestRun(t *testing.T, opts TestRunCreateOptions) *Run {
	org, err := NewOrganization(OrganizationCreateOptions{Name: String("test-org")})
	require.NoError(t, err)

	ws, err := NewWorkspace(org, WorkspaceCreateOptions{Name: "test-ws", AutoApply: Bool(opts.AutoApply)})
	require.NoError(t, err)

	cv, err := NewConfigurationVersion(ws.ID(), ConfigurationVersionCreateOptions{Speculative: Bool(opts.Speculative)})
	require.NoError(t, err)

	run := NewRun(cv, ws, RunCreateOptions{})
	if opts.Status != RunStatus("") {
		run.updateStatus(opts.Status)
	}
	if opts.ID != nil {
		run.id = *opts.ID
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
