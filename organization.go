package otf

import (
	"context"
)

var (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	id string `json:"organization_id"`

	Timestamps

	name            string `json:"name"`
	sessionRemember int    `json:"session_remember"`
	sessionTimeout  int    `json:"session_timeout"`
}

// OrganizationCreateOptions represents the options for creating an
// organization.
type OrganizationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// Name of the organization.
	Name *string `jsonapi:"attr,name"`

	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`
}

// OrganizationUpdateOptions represents the options for updating an
// organization.
type OrganizationUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// New name for the organization.
	Name *string `jsonapi:"attr,name,omitempty"`

	// Session expiration (minutes).
	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`
}

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
	Create(org *Organization) (*Organization, error)
	Get(name string) (*Organization, error)
	List(opts OrganizationListOptions) (*OrganizationList, error)
	Update(name string, fn func(*Organization) error) (*Organization, error)
	Delete(name string) error
}

func (org *Organization) ID() string           { return org.id }
func (org *Organization) String() string       { return org.id }
func (org *Organization) Name() string         { return org.name }
func (org *Organization) SessionRemember() int { return org.sessionRemember }
func (org *Organization) SessionTimeout() int  { return org.sessionTimeout }

func (o OrganizationCreateOptions) Valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

func NewOrganization(opts OrganizationCreateOptions) (*Organization, error) {
	org := Organization{
		name:            *opts.Name,
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
