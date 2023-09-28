package repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name     string
		hostname string
		opts     newHookOptions
		want     *Hook
	}{
		{
			name:     "default",
			hostname: "fakehost.org",
			opts: newHookOptions{
				id:     &id,
				cloud:  cloud.GithubKind,
				secret: internal.String("top-secret"),
			},
			want: &Hook{
				id:           id,
				secret:       "top-secret",
				cloud:        cloud.GithubKind,
				endpoint:     "https://fakehost.org/webhooks/vcs/" + id.String(),
				cloudHandler: gitlab.EventHandler{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory{
				HostnameService: internal.NewHostnameService(tt.hostname),
			}
			got, err := f.newHook(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
