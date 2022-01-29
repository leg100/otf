package html

import (
	"context"

	"github.com/leg100/otf"
)

type fakeUserService struct {
	otf.UserService
	database map[string]*otf.User
}

func (s *fakeUserService) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	user, ok := s.database[*spec.Username]
	if !ok {
		return nil, otf.ErrResourceNotFound
	}
	return user, nil
}

func (s *fakeUserService) Create(ctx context.Context, name string) (*otf.User, error) {
	u := &otf.User{Username: name}
	s.database[name] = u
	return u, nil
}

func (s *fakeUserService) Update(ctx context.Context, name string, updated *otf.User) error {
	s.database[name] = updated
	return nil
}

type fakeOrganizationService struct {
	otf.OrganizationService
	database map[string]*otf.Organization
}

func (s *fakeOrganizationService) Get(ctx context.Context, name string) (*otf.Organization, error) {
	org, ok := s.database[name]
	if !ok {
		return nil, otf.ErrResourceNotFound
	}
	return org, nil
}

func (s *fakeOrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	o := &otf.Organization{Name: *opts.Name}
	s.database[*opts.Name] = o
	return o, nil
}
