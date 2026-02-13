package repohooks

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newHook(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name string
		opts newRepohookOptions
		want *hook
	}{
		{
			name: "default",
			opts: newRepohookOptions{
				id:              &id,
				vcsKindID:       vcs.KindID("test"),
				secret:          new("top-secret"),
				HostnameService: internal.NewHostnameService("fakehost.org"),
			},
			want: &hook{
				id:        id,
				secret:    "top-secret",
				vcsKindID: vcs.KindID("test"),
				endpoint:  "https://fakehost.org/webhooks/vcs/" + id.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRepohook(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
