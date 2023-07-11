package repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name     string
		hostname string
		opts     newHookOptions
		want     *hook
	}{
		{
			name:     "default",
			hostname: "fakehost.org",
			opts: newHookOptions{
				id:     &id,
				cloud:  "fakecloud",
				secret: internal.String("top-secret"),
			},
			want: &hook{
				id:           id,
				secret:       "top-secret",
				cloud:        "fakecloud",
				endpoint:     "https://fakehost.org/webhooks/vcs/" + id.String(),
				EventHandler: &fakeCloud{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory{
				HostnameService: fakeHostnameService{hostname: tt.hostname},
				Service:         fakeCloudService{},
			}
			got, err := f.newHook(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
