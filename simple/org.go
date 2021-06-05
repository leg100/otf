package simple

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
)

type OrganizationService map[string]*tfe.Organization

func NewOrganizationService() OrganizationService {
	return OrganizationService(make(map[string]*tfe.Organization))
}

func (s OrganizationService) CreateOrganization(name string, org *tfe.Organization) (*tfe.Organization, error) {
	if _, ok := s[name]; ok {
		return nil, fmt.Errorf("already exists")
	}

	s[name] = org

	return org, nil
}

func (s OrganizationService) GetOrganization(name string) (*tfe.Organization, error) {
	org, ok := s[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return org, nil
}

func (s OrganizationService) ListOrganizations() ([]*tfe.Organization, error) {
	var orgs []*tfe.Organization
	for _, o := range s {
		orgs = append(orgs, o)
	}

	return orgs, nil
}

func (s OrganizationService) UpdateOrganization(name string, org *tfe.Organization) (*tfe.Organization, error) {
	if _, ok := s[name]; !ok {
		return nil, fmt.Errorf("not found")
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
func (s OrganizationService) GetEntitlements(name string) (*tfe.Entitlements, error) {
	if _, ok := s[name]; !ok {
		return nil, fmt.Errorf("not found")
	}

	return &tfe.Entitlements{}, nil
}
