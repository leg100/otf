package simple

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
)

type OrganizationService map[string]*ots.Organization

func NewOrganizationService(orgs ...*ots.Organization) OrganizationService {
	db := make(map[string]*ots.Organization)
	for _, org := range orgs {
		db[org.Name] = org
	}
	return OrganizationService(db)
}

func (s OrganizationService) CreateOrganization(name string, org *ots.Organization) (*ots.Organization, error) {
	if _, ok := s[name]; ok {
		return nil, fmt.Errorf("already exists")
	}

	s[name] = org

	return org, nil
}

func (s OrganizationService) GetOrganization(name string) (*ots.Organization, error) {
	org, ok := s[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return org, nil
}

func (s OrganizationService) ListOrganizations() ([]*ots.Organization, error) {
	var orgs []*ots.Organization
	for _, o := range s {
		orgs = append(orgs, o)
	}

	return orgs, nil
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
	org := &ots.Organization{}

	if _, ok := s[name]; !ok {
		return nil, fmt.Errorf("not found")
	}

	if err := ots.UpdateOrganizationFromOptions(org, opts); err != nil {
		return nil, err
	}

	s[name] = org

	return org, nil
}

func (s OrganizationService) DeleteOrganization(name string) error {
	if _, ok := s[name]; !ok {
		return fmt.Errorf("not found")
	}

	delete(s, name)

	return nil
}

// GetEntitlements currently shows all entitlements as disabled for an org.
func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	if _, ok := s[name]; !ok {
		return nil, fmt.Errorf("not found")
	}

	return &ots.Entitlements{}, nil
}
