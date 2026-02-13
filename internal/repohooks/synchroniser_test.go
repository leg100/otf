package repohooks

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSynchroniser tests synchronising a hook with a cloud provider's hook,
// seeding the cloud with a different state in each test case.
func TestSynchroniser(t *testing.T) {
	tests := []struct {
		name  string
		cloud vcs.Webhook // seed cloud with hook
		got   *hook       // seed db with hook
		want  *hook       // hook after synchronisation
	}{
		{
			name: "synchronised",
			cloud: vcs.Webhook{
				ID:       "123",
				Events:   defaultEvents,
				Endpoint: "fake-host.org/xyz",
			},
			got: &hook{
				cloudID:  new("123"),
				endpoint: "fake-host.org/xyz",
			},
			want: &hook{
				cloudID:  new("123"),
				endpoint: "fake-host.org/xyz",
			},
		},
		{
			name:  "new hook",
			cloud: vcs.Webhook{ID: "123"}, // new id that cloud returns
			got: &hook{
				endpoint: "fake-host.org/xyz",
			},
			want: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  new("123"),
			},
		},
		{
			name: "hook events missing on cloud",
			cloud: vcs.Webhook{
				ID:       "123",
				Endpoint: "fake-host.org/xyz",
			},
			got: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  new("123"),
			},
			want: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  new("123"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeCloudClient{hook: tt.cloud}
			db := &fakeDB{hook: tt.got}
			synchr := &synchroniser{Logger: logr.Discard(), syncdb: db}
			require.NoError(t, synchr.sync(context.Background(), client, tt.got))
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
