package organization

import (
	"context"

	"github.com/leg100/otf/internal/apigen"
)

type api struct {
	Service
}

func (a *api) GetOrganization(ctx context.Context, params apigen.GetOrganizationParams) (apigen.GetOrganizationRes, error) {
	org, err := a.Service.GetOrganization(ctx, params.Name)
	if err != nil {
		return nil, err
	}
	return &apigen.Organization{ID: org.ID, Name: org.Name}, nil
}

func (a *api) DeleteOrganization(ctx context.Context, params apigen.DeleteOrganizationParams) error {
	return a.Service.DeleteOrganization(ctx, params.Name)
}
