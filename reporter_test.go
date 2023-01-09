package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReporter checks the scheduler is creating workspace queues and
// forwarding events to the queue handlers.
func TestReporter_HandleRun(t *testing.T) {
	reporter, run, updates := newTestReporter(t, RunPending)

	err := reporter.handleRun(context.Background(), run)
	require.NoError(t, err)
	got := <-updates
	assert.Equal(t, VCSPendingStatus, got.Status)
}

// newTestReporter creates a reporter for testing purposes, returning a run
// with the given status and a channel of status updates.
func newTestReporter(t *testing.T, status RunStatus) (*Reporter, *Run, <-chan SetStatusOptions) {
	org := NewTestOrganization(t)
	provider := NewTestVCSProvider(t, org)
	hook := NewTestWebhook(NewTestRepo(), NewTestCloudConfig(nil))
	repo := NewTestWorkspaceRepo(provider, hook)
	ws := NewTestWorkspace(t, org, WithRepo(repo))
	cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{
		IngressAttributes: &IngressAttributes{},
	})

	run := NewRun(cv, ws, RunCreateOptions{})
	run.status = status
	statusUpdates := make(chan SetStatusOptions, 1)
	reporter := NewReporter(logr.Discard(), &fakeReporterApp{
		ws:            ws,
		cv:            cv,
		statusUpdates: statusUpdates,
		hostname:      "fake-host.org",
	})
	return reporter, run, statusUpdates
}

type fakeReporterApp struct {
	hostname      string
	ws            *Workspace
	cv            *ConfigurationVersion
	statusUpdates chan SetStatusOptions

	Application
}

func (f *fakeReporterApp) Hostname() string {
	return f.hostname
}

func (f *fakeReporterApp) GetWorkspace(context.Context, WorkspaceSpec) (*Workspace, error) {
	return f.ws, nil
}

func (f *fakeReporterApp) GetConfigurationVersion(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeReporterApp) SetStatus(ctx context.Context, id string, opts SetStatusOptions) error {
	f.statusUpdates <- opts
	return nil
}
