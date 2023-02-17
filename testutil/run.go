package testutil

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

// TestRunCreateOptions is for testing purposes only.
type TestRunCreateOptions struct {
	ID                *string // override ID of run
	Speculative       bool
	ExecutionMode     *otf.ExecutionMode
	Status            otf.RunStatus
	AutoApply         *bool
	Repo              *otf.WorkspaceRepo
	IngressAttributes *otf.IngressAttributes
	Workspace         otf.Workspace // run's workspace; if nil a workspace is auto created
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	ctx := context.Background()
	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	err := db.CreateRun(ctx, run)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteRun(ctx, run.ID())
	})
	return run
}

// NewTestRun creates a new run. Expressly for testing purposes.
func NewTestRun(t *testing.T, opts TestRunCreateOptions) *Run {
	org := organization.NewTestOrganization(t)

	ws := opts.Workspace
	if ws == nil {
		ws, err := NewWorkspace(CreateWorkspaceOptions{
			Name:         String("test-ws"),
			Organization: String(org.Name()),
			Repo:         opts.Repo,
		})
		require.NoError(t, err)
	}

	cv, err := NewConfigurationVersion(ws.ID(), ConfigurationVersionCreateOptions{
		IngressAttributes: opts.IngressAttributes,
		Speculative:       otf.Bool(opts.Speculative),
	})
	require.NoError(t, err)

	run := NewRun(cv, ws, RunCreateOptions{
		AutoApply: opts.AutoApply,
	})

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

func (f *fakeRunFactoryWorkspaceService) GetWorkspace(context.Context, string) (*Workspace, error) {
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
