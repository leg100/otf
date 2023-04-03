package orgcreator

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name     string
		subject  otf.Subject
		restrict bool
		want     error
	}{
		{"site admin", &auth.SiteAdmin, false, nil},
		{"normal user", &auth.User{}, false, nil},
		{"non-user", &auth.AgentToken{}, false, otf.ErrAccessNotPermitted},
		{"restrict to site admin - site admin", &auth.SiteAdmin, true, nil},
		{"restrict to site admin - user", &auth.User{}, true, otf.ErrAccessNotPermitted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := otf.AddSubjectToContext(context.Background(), tt.subject)
			svc := &service{
				Logger:                       logr.Discard(),
				RestrictOrganizationCreation: tt.restrict,
			}
			_, err := svc.authorize(ctx)
			assert.Equal(t, tt.want, err)
		})
	}
}
