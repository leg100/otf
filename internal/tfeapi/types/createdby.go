package api

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/user"
)

// CreatedBy provides the creator of a TFE resource. Only one value can be
// non-nil.
type CreatedBy struct {
	Organization *organization.TFEOrganization
	User         *user.TFEUser
}

func NewCreatedBy(ctx context.Context) (*CreatedBy, error) {
	subj, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	switch creator := subj.(type) {
	case *user.User:
		return &CreatedBy{User: &user.TFEUser{}}, nil
	case *organization.OrganizationToken:
		return &CreatedBy{Organization: &organization.TFEOrganization{
			Name: creator.Organization,
		}}, nil
	default:
		return nil, fmt.Errorf("unexpected creator: %T", creator)
	}
}
