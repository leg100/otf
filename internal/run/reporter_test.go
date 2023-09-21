package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter_HandleRun(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  *Run
		ws   *workspace.Workspace
		cv   *configversion.ConfigurationVersion
		want cloud.SetStatusOptions
	}{
		{
			name: "pending run",
			run:  &Run{ID: "run-123", Status: internal.RunPending},
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
			want: cloud.SetStatusOptions{
				Workspace: "dev",
				Ref:       "abc123",
				Repo:      "leg100/otf",
				Status:    cloud.VCSPendingStatus,
				TargetURL: "https://otf-host.org/app/runs/run-123",
			},
		},
		{
			name: "skip run with config not from a VCS repo",
			run:  &Run{ID: "run-123"},
			cv: &configversion.ConfigurationVersion{
				IngressAttributes: nil,
			},
			want: cloud.SetStatusOptions{},
		},
		{
			name: "skip UI-triggered run",
			run:  &Run{ID: "run-123", Source: SourceUI},
			want: cloud.SetStatusOptions{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got cloud.SetStatusOptions
			reporter := &Reporter{
				WorkspaceService:            &fakeReporterWorkspaceService{ws: tt.ws},
				ConfigurationVersionService: &fakeReporterConfigurationVersionService{cv: tt.cv},
				VCSProviderService:          &fakeReporterVCSProviderService{got: &got},
				HostnameService:             internal.NewHostnameService("otf-host.org"),
			}
			err := reporter.handleRun(ctx, tt.run)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeReporterConfigurationVersionService struct {
	configversion.Service

	cv *configversion.ConfigurationVersion
}

func (f *fakeReporterConfigurationVersionService) GetConfigurationVersion(context.Context, string) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

type fakeReporterWorkspaceService struct {
	workspace.Service

	ws *workspace.Workspace
}

func (f *fakeReporterWorkspaceService) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

type fakeReporterVCSProviderService struct {
	vcsprovider.VCSProviderService

	got *cloud.SetStatusOptions
}

func (f *fakeReporterVCSProviderService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeReporterCloudClient{got: f.got}, nil
}

type fakeReporterCloudClient struct {
	cloud.Client

	got *cloud.SetStatusOptions
}

func (f *fakeReporterCloudClient) SetStatus(ctx context.Context, opts cloud.SetStatusOptions) error {
	*f.got = opts
	return nil
}
