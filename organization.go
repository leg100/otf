package otf

import (
	"context"
)

var (
	DefaultSessionTimeout          = 20160
	DefaultSessionExpiration       = 20160
	DefaultOrganizationPermissions = OrganizationPermissions{
		CanCreateWorkspace: true,
		CanUpdate:          true,
		CanDestroy:         true,
	}
)

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	ID string `json:"organization_id"`

	Timestamps

	Name            string `json:"name"`
	SessionRemember int    `json:"session_remember"`
	SessionTimeout  int    `json:"session_timeout"`
}

// OrganizationPermissions represents the organization permissions.
type OrganizationPermissions struct {
	CanCreateTeam               bool `json:"can-create-team"`
	CanCreateWorkspace          bool `json:"can-create-workspace"`
	CanCreateWorkspaceMigration bool `json:"can-create-workspace-migration"`
	CanDestroy                  bool `json:"can-destroy"`
	CanTraverse                 bool `json:"can-traverse"`
	CanUpdate                   bool `json:"can-update"`
	CanUpdateAPIToken           bool `json:"can-update-api-token"`
	CanUpdateOAuth              bool `json:"can-update-oauth"`
	CanUpdateSentinel           bool `json:"can-update-sentinel"`
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

func (org *Organization) GetID() string  { return org.ID }
func (org *Organization) String() string { return org.ID }

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
		Name:            *opts.Name,
		ID:              NewID("org"),
		SessionTimeout:  DefaultSessionTimeout,
		SessionRemember: DefaultSessionExpiration,
	}

	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}

	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}

	return &org, nil
}

func UpdateOrganizationFromOpts(org *Organization, opts OrganizationUpdateOptions) error {
	if opts.Name != nil {
		org.Name = *opts.Name
	}

	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}

	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}

	return nil
}
