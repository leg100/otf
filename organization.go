package otf

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	id              string
	createdAt       time.Time
	updatedAt       time.Time
	name            string
	sessionRemember int
	sessionTimeout  int
}

func (org *Organization) ID() string               { return org.id }
func (org *Organization) CreatedAt() time.Time     { return org.createdAt }
func (org *Organization) UpdatedAt() time.Time     { return org.updatedAt }
func (org *Organization) String() string           { return org.id }
func (org *Organization) Name() string             { return org.name }
func (org *Organization) OrganizationName() string { return org.name }
func (org *Organization) SessionRemember() int     { return org.sessionRemember }
func (org *Organization) SessionTimeout() int      { return org.sessionTimeout }

// OrganizationList represents a list of Organizations.
type OrganizationList struct {
	*Pagination
	Items []*Organization
}

// OrganizationListOptions represents the options for listing organizations.
type OrganizationListOptions struct {
	ListOptions
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
	EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
	GetOrganization(ctx context.Context, name string) (*Organization, error)
	ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
	UpdateOrganization(ctx context.Context, name string, opts *OrganizationUpdateOptions) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error
	GetEntitlements(ctx context.Context, name string) (*Entitlements, error)
}

type OrganizationStore interface {
	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, name string) (*Organization, error)
	ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
	UpdateOrganization(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error
	GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error)
}

// OrganizationCreateOptions represents the options for creating an
// organization. See dto.OrganizationCreateOptions for more details.
type OrganizationCreateOptions struct {
	Name            *string `schema:"name,required"`
	SessionRemember *int
	SessionTimeout  *int
}

func (opts *OrganizationCreateOptions) Validate() error {
	if opts.Name == nil {
		return errors.New("name required")
	}
	if *opts.Name == "" {
		return errors.New("name cannot be empty")
	}
	if !ValidStringID(opts.Name) {
		return fmt.Errorf("invalid name: %s", *opts.Name)
	}
	return nil
}

// OrganizationUpdateOptions represents the options for updating an
// organization. See dto.OrganizationUpdateOptions for more details.
type OrganizationUpdateOptions struct {
	Name            *string
	SessionRemember *int
	SessionTimeout  *int
}

func NewOrganization(opts OrganizationCreateOptions) (*Organization, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	org := Organization{
		name:            *opts.Name,
		createdAt:       CurrentTimestamp(),
		updatedAt:       CurrentTimestamp(),
		id:              NewID("org"),
		sessionTimeout:  DefaultSessionTimeout,
		sessionRemember: DefaultSessionExpiration,
	}
	if opts.SessionTimeout != nil {
		org.sessionTimeout = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.sessionRemember = *opts.SessionRemember
	}
	return &org, nil
}

func UpdateOrganizationFromOpts(org *Organization, opts OrganizationUpdateOptions) error {
	if opts.Name != nil {
		org.name = *opts.Name
	}
	if opts.SessionTimeout != nil {
		org.sessionTimeout = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.sessionRemember = *opts.SessionRemember
	}
	return nil
}
