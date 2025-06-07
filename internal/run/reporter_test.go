package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter_HandleRun(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		event *Event
		ws    *workspace.Workspace
		cv    *configversion.ConfigurationVersion
		// expect the given status options to be set. If nil then expect no
		// status options to be set.
		want *vcs.SetStatusOptions
	}{
		{
			name:  "set pending status",
			event: &Event{ID: testutils.ParseID(t, "run-123"), Status: runstatus.Pending},
			ws: &workspace.Workspace{
				Name:       "dev",
				Connection: &workspace.Connection{},
			},
			cv: &configversion.ConfigurationVersion{
				IngressAttributes: &configversion.IngressAttributes{
					CommitSHA: "abc123",
					Repo:      "leg100/otf",
				},
			},
			want: &vcs.SetStatusOptions{
				Workspace: "dev",
				Ref:       "abc123",
				Repo:      "leg100/otf",
				Status:    vcs.PendingStatus,
				TargetURL: "https://otf-host.org/app/runs/run-123",
			},
		},
		{
			name:  "skip run with config not from a VCS repo",
			event: &Event{ID: testutils.ParseID(t, "run-123")},
			cv: &configversion.ConfigurationVersion{
				IngressAttributes: nil,
			},
			want: nil,
		},
		{
			name:  "skip UI-triggered run",
			event: &Event{ID: testutils.ParseID(t, "run-123"), Source: SourceUI},
			want:  nil,
		},
		{
			name:  "skip API-triggered run",
			event: &Event{ID: testutils.ParseID(t, "run-123"), Source: SourceAPI},
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(chan vcs.SetStatusOptions, 1)
			reporter := &Reporter{
				Workspaces:      &fakeReporterWorkspaceClient{ws: tt.ws},
				Runs:            &fakeReporterRunClient{event: tt.event},
				Configs:         &fakeReporterConfigurationVersionService{cv: tt.cv},
				VCS:             &fakeReporterVCSProviderService{got: got},
				HostnameService: internal.NewHostnameService("otf-host.org"),
				Cache:           make(map[resource.TfeID]vcs.Status),
			}
			err := reporter.handleRun(ctx, tt.event)
			require.NoError(t, err)

			if tt.want == nil {
				assert.Equal(t, 0, len(got))
			} else {
				assert.Equal(t, *tt.want, <-got)
			}
		})
	}
}

// TestReporter_DontSetStatusTwice tests that the same status is not set more
// than once for a given run.
func TestReporter_DontSetStatusTwice(t *testing.T) {
	ctx := context.Background()

	event := &Event{ID: testutils.ParseID(t, "run-123"), Status: runstatus.Pending}
	ws := &workspace.Workspace{
		Name:       "dev",
		Connection: &workspace.Connection{},
	}
	cv := &configversion.ConfigurationVersion{
		IngressAttributes: &configversion.IngressAttributes{
			CommitSHA: "abc123",
			Repo:      "leg100/otf",
		},
	}

	got := make(chan vcs.SetStatusOptions, 1)
	reporter := &Reporter{
		Workspaces:      &fakeReporterWorkspaceClient{ws: ws},
		Configs:         &fakeReporterConfigurationVersionService{cv: cv},
		VCS:             &fakeReporterVCSProviderService{got: got},
		Runs:            &fakeReporterRunClient{event: event},
		HostnameService: internal.NewHostnameService("otf-host.org"),
		Cache:           make(map[resource.TfeID]vcs.Status),
	}

	// handle run the first time and expect status to be set
	err := reporter.handleRun(ctx, event)
	require.NoError(t, err)

	want := vcs.SetStatusOptions{
		Workspace: "dev",
		Ref:       "abc123",
		Repo:      "leg100/otf",
		Status:    vcs.PendingStatus,
		TargetURL: "https://otf-host.org/app/runs/run-123",
	}
	assert.Equal(t, want, <-got)

	// handle run the second time with the same status and expect status to
	// *not* be set
	err = reporter.handleRun(ctx, event)
	require.NoError(t, err)
	assert.Equal(t, 0, len(got))
}

type fakeReporterConfigurationVersionService struct {
	configversion.Service

	cv *configversion.ConfigurationVersion
}

func (f *fakeReporterConfigurationVersionService) Get(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

type fakeReporterWorkspaceClient struct {
	workspace.Service

	ws *workspace.Workspace
}

func (f *fakeReporterWorkspaceClient) Get(context.Context, resource.TfeID) (*workspace.Workspace, error) {
	return f.ws, nil
}

type fakeReporterRunClient struct {
	reporterRunClient

	event *Event
}

func (f *fakeReporterRunClient) Get(context.Context, resource.TfeID) (*Run, error) {
	return &Run{ID: f.event.ID, Status: f.event.Status}, nil
}

type fakeReporterVCSProviderService struct {
	got chan vcs.SetStatusOptions
}

func (f *fakeReporterVCSProviderService) Get(context.Context, resource.TfeID) (*vcs.Provider, error) {
	return &vcs.Provider{
		Client: &fakeReporterCloudClient{got: f.got},
	}, nil
}

type fakeReporterCloudClient struct {
	vcs.Client

	got chan vcs.SetStatusOptions
}

func (f *fakeReporterCloudClient) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	f.got <- opts
	return nil
}
