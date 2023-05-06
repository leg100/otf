package repo

import (
	"testing"

	"github.com/google/uuid"
	internal "github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name     string
		hostname string
		opts     newHookOpts
		want     *hook
	}{
		{
			name:     "default",
			hostname: "fakehost.org",
			opts: newHookOpts{
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
			f := newFactory(
				fakeHostnameService{hostname: tt.hostname},
				fakeCloudService{},
			)
			got, err := f.newHook(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
