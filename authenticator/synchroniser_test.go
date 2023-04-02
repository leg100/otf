package authenticator

import (
	"testing"

	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
)

func TestTeamDiff(t *testing.T) {
	tests := []struct {
		name       string
		want       []*auth.Team
		got        []*auth.Team
		wantAdd    []*auth.Team
		wantRemove []*auth.Team
	}{
		{
			name:    "new user",
			want:    []*auth.Team{{Name: "owners", Organization: "acme-corp"}},
			wantAdd: []*auth.Team{{Name: "owners", Organization: "acme-corp"}},
		},
		{
			name: "existing user",
			want: []*auth.Team{{Name: "owners", Organization: "acme-corp"}},
			got:  []*auth.Team{{Name: "owners", Organization: "acme-corp"}},
		},
		{
			name: "no longer a member of devs team",
			want: []*auth.Team{{Name: "owners", Organization: "acme-corp"}},
			got: []*auth.Team{
				{Name: "owners", Organization: "acme-corp"},
				{Name: "devs", Organization: "acme-corp"},
			},
			wantRemove: []*auth.Team{{Name: "devs", Organization: "acme-corp"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdd, gotRemove := teamdiff(tt.want, tt.got)
			assert.Equal(t, tt.wantAdd, gotAdd)
			assert.Equal(t, tt.wantRemove, gotRemove)
		})
	}
}
