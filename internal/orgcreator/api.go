package orgcreator

import (
	"context"

	"github.com/leg100/otf/internal/apigen"
)

type api struct {
	Service
}

func (a *api) CreateOrganization(ctx context.Context, req *apigen.NewOrganization) (*apigen.Organization, error) {
	org, err := a.Service.CreateOrganization(ctx, OrganizationCreateOptions{
		Name: &req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &apigen.Organization{ID: org.ID, Name: org.Name}, nil
}
