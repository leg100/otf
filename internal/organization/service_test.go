package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name     string
		subject  internal.Subject
		restrict bool
		want     error
	}{
		{"site admin", &internal.Superuser{}, false, nil},
		{"restrict to site admin - site admin", &internal.Superuser{}, true, nil},
		{"restrict to site admin - user", &unprivUser{}, true, internal.ErrAccessNotPermitted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := internal.AddSubjectToContext(context.Background(), tt.subject)
			svc := &service{
				Logger:                       logr.Discard(),
				RestrictOrganizationCreation: tt.restrict,
			}
			_, err := svc.restrictOrganizationCreation(ctx)
			assert.Equal(t, tt.want, err)
		})
	}
}

type unprivUser struct {
	internal.Subject
}

func (s *unprivUser) IsSiteAdmin() bool { return false }
