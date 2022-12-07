package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunFactory(t *testing.T) {
	org := NewTestOrganization(t)
	ws := NewTestWorkspace(t, org)
	autoApplyWS := NewTestWorkspace(t, org, AutoApply())
	cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
	speculativeCV := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{
		Speculative: Bool(true),
	})

	tests := []struct {
		name string
		opts RunCreateOptions
		want func(*testing.T, *Run, error)
		ws   *Workspace
		cv   *ConfigurationVersion
	}{
		{
			name: "defaults",
			ws:   ws,
			cv:   cv,
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.Equal(t, RunPending, run.status)
				assert.NotZero(t, run.createdAt)
				assert.False(t, run.speculative)
				assert.True(t, run.refresh)
				assert.False(t, run.autoApply)
			},
		},
		{
			name: "speculative",
			ws:   ws,
			cv:   speculativeCV,
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.True(t, run.speculative)
			},
		},
		{
			name: "auto-apply",
			ws:   autoApplyWS,
			cv:   cv,
			opts: RunCreateOptions{},
			want: func(t *testing.T, run *Run, err error) {
				assert.True(t, run.autoApply)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := RunFactory{
				WorkspaceService:            &fakeRunFactoryWorkspaceService{ws: tt.ws},
				ConfigurationVersionService: &fakeRunFactoryConfigurationVersionService{cv: tt.cv},
			}
			run, err := f.NewRun(context.Background(), tt.ws.SpecID(), tt.opts)
			tt.want(t, run, err)
		})
	}
}
