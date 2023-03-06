package repo

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHook_Sync tests synchronising a hook with a cloud provider's hook,
// seeding the cloud with a different state in each test case.
func TestHook_Sync(t *testing.T) {
	tests := []struct {
		name       string
		cloud      cloud.Webhook // seed cloud with hook
		got        *hook
		want       *hook // hook after synchronisation
		wantUpdate bool  // want cloud to be updated
	}{
		{
			name: "synchronised",
			cloud: cloud.Webhook{
				ID:       "123",
				Events:   defaultEvents,
				Endpoint: "fake-host.org/xyz",
			},
			got: &hook{
				cloudID:  otf.String("123"),
				endpoint: "fake-host.org/xyz",
			},
			want: &hook{
				cloudID:  otf.String("123"),
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
				cloudID:  otf.String("123"),
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
				cloudID:  otf.String("123"),
			},
			want: &hook{
				endpoint: "fake-host.org/xyz",
				cloudID:  otf.String("123"),
			},
			wantUpdate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeCloudClient{hook: tt.cloud}
			err := tt.got.sync(context.Background(), client)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.got)
			assert.Equal(t, tt.wantUpdate, client.gotUpdate)
		})
	}
}
