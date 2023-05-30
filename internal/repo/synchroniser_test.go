package repo

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSynchroniser tests synchronising a hook with a cloud provider's hook,
// seeding the cloud with a different state in each test case.
func TestSynchroniser(t *testing.T) {
	tests := []struct {
		name  string
		cloud cloud.Webhook // seed cloud with hook
		got   *hook         // seed db with hook
		want  *hook         // hook after synchronisation
	}{
		{
			name: "synchronised",
			cloud: cloud.Webhook{
				ID:       "123",
				Events:   defaultEvents,
				Endpoint: "fake-host.org/xyz",
			},
			got: &hook{
				cloudID:  internal.String("123"),
				endpoint: "fake-host.org/xyz",
			},
			want: &hook{
				cloudID:  internal.String("123"),
				endpoint: "fake-host.org/xyz",
			},
		},
		{
			name:  "new hook",
			cloud: cloud.Webhook{ID: "123"}, // new id that cloud returns
			got: &hook{
				endpoint: "fake-host.org/xyz",
			},
			want: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  internal.String("123"),
			},
		},
		{
			name: "hook events missing on cloud",
			cloud: cloud.Webhook{
				ID:       "123",
				Endpoint: "fake-host.org/xyz",
			},
			got: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  internal.String("123"),
			},
			want: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  internal.String("123"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeCloudClient{hook: tt.cloud}
			db := &fakeDB{hook: tt.got}
			synchr := &synchroniser{Logger: logr.Discard()}
			require.NoError(t, synchr.sync(context.Background(), db, client, tt.got))
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
