package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name     string
		subject  authz.Subject
		restrict bool
		want     error
	}{
		{"site admin", &authz.Superuser{}, false, nil},
		{"restrict to site admin - site admin", &authz.Superuser{}, true, nil},
		{"restrict to site admin - user", &unprivUser{}, true, internal.ErrAccessNotPermitted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authz.AddSubjectToContext(context.Background(), tt.subject)
			svc := &Service{
				Logger:                       logr.Discard(),
				RestrictOrganizationCreation: tt.restrict,
			}
			_, err := svc.restrictOrganizationCreation(ctx)
			assert.Equal(t, tt.want, err)
		})
	}
}

type unprivUser struct {
	authz.Subject
}

func (s *unprivUser) IsSiteAdmin() bool { return false }
