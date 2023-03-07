package run

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter_HandleRun(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  *otf.Run
		ws   *otf.Workspace
		cv   *otf.ConfigurationVersion
		want cloud.SetStatusOptions
	}{
		{
			name: "pending run",
			run:  &otf.Run{Status: otf.RunPending},
			ws: &otf.Workspace{
				Name: "dev",
				Repo: &otf.Connection{Identifier: "leg100/otf"},
			},
			cv: &otf.ConfigurationVersion{
				IngressAttributes: &otf.IngressAttributes{CommitSHA: "abc123"},
			},
			want: cloud.SetStatusOptions{
				Workspace:  "ws-123",
				Ref:        "abc123",
				Identifier: "leg100/otf",
				Status:     cloud.VCSPendingStatus,
				TargetURL:  "https://otf-host.org/runs/run-123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(chan cloud.SetStatusOptions, 1)
			reporter := &reporter{
				WorkspaceService:            &fakeReporterWorkspaceService{ws: tt.ws},
				ConfigurationVersionService: &fakeReporterConfigurationVersionService{cv: tt.cv},
				VCSProviderService:          &fakeReporterVCSProviderService{got: got},
				hostname:                    "otf-host.org",
			}
			err := reporter.handleRun(ctx, tt.run)
			require.NoError(t, err)

			assert.Equal(t, tt.want, <-got)
		})
	}
}

type fakeReporterConfigurationVersionService struct {
	otf.ConfigurationVersionService

	cv *otf.ConfigurationVersion
}

func (f *fakeReporterConfigurationVersionService) GetConfigurationVersion(context.Context, string) (*otf.ConfigurationVersion, error) {
	return f.cv, nil
}

type fakeReporterWorkspaceService struct {
	otf.WorkspaceService

	ws *otf.Workspace
}

func (f *fakeReporterWorkspaceService) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.ws, nil
}

type fakeReporterVCSProviderService struct {
	otf.VCSProviderService

	got chan cloud.SetStatusOptions
}

func (f *fakeReporterVCSProviderService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeReporterCloudClient{got: f.got}, nil
}

type fakeReporterCloudClient struct {
	cloud.Client

	got chan cloud.SetStatusOptions
}

func (f *fakeReporterCloudClient) SetStatus(ctx context.Context, opts cloud.SetStatusOptions) error {
	f.got <- opts
	return nil
}
