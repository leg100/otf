package http

import (
	"context"

	"github.com/leg100/otf"
)

// Compile-time proof of interface implementation.
var _ otf.OrganizationService = (*organizations)(nil)

// organizations implements OrganizationService.
type organizations struct {
	client *client

	// TODO: implement all of otf.OrganizationService's methods
	otf.OrganizationService
}

// Create a new organization with the given options.
func (s *organizations) Create(ctx context.Context, options otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("POST", "organizations", &options)
	if err != nil {
		return nil, err
	}

	org := &Organization{}
	err = s.client.do(ctx, req, org)
	if err != nil {
		return nil, err
	}

	return org.ToDomain(), nil
}
