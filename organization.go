package otf

import (
	"context"
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

func (org *Organization) ID() string           { return org.id }
func (org *Organization) CreatedAt() time.Time { return org.createdAt }
func (org *Organization) UpdatedAt() time.Time { return org.updatedAt }
func (org *Organization) String() string       { return org.id }
func (org *Organization) Name() string         { return org.name }
func (org *Organization) SessionRemember() int { return org.sessionRemember }
func (org *Organization) SessionTimeout() int  { return org.sessionTimeout }

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
	Create(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
	EnsureCreated(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
	Get(ctx context.Context, name string) (*Organization, error)
	List(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
	Update(ctx context.Context, name string, opts *OrganizationUpdateOptions) (*Organization, error)
	Delete(ctx context.Context, name string) error
	GetEntitlements(ctx context.Context, name string) (*Entitlements, error)
}

type OrganizationStore interface {
	Create(org *Organization) error
	Get(name string) (*Organization, error)
	List(opts OrganizationListOptions) (*OrganizationList, error)
	Update(name string, fn func(*Organization) error) (*Organization, error)
	Delete(name string) error
}

// OrganizationCreateOptions represents the options for creating an
// organization. See dto.OrganizationCreateOptions for more details.
type OrganizationCreateOptions struct {
	Name            *string
	SessionRemember *int
	SessionTimeout  *int
}

func (opts *OrganizationCreateOptions) Validate() error {
	if !validString(opts.Name) {
		return ErrRequiredName
	}
	if !ValidStringID(opts.Name) {
		return ErrInvalidName
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
