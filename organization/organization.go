package organization

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf"
)

var (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

// Organization is an OTF Organization, comprising workspaces, users, etc.
type Organization struct {
	id              string
	createdAt       time.Time
	updatedAt       time.Time
	name            string
	sessionRemember int
	sessionTimeout  int
}

func NewOrganization(opts OrganizationCreateOptions) (*Organization, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	org := Organization{
		name:            *opts.Name,
		createdAt:       otf.CurrentTimestamp(),
		updatedAt:       otf.CurrentTimestamp(),
		id:              otf.NewID("org"),
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

func (org *Organization) ID() string           { return org.id }
func (org *Organization) CreatedAt() time.Time { return org.createdAt }
func (org *Organization) UpdatedAt() time.Time { return org.updatedAt }
func (org *Organization) String() string       { return org.id }
func (org *Organization) Name() string         { return org.name }
func (org *Organization) SessionRemember() int { return org.sessionRemember }
func (org *Organization) SessionTimeout() int  { return org.sessionTimeout }

func (org *Organization) Update(opts OrganizationUpdateOptions) error {
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

// ToJSONAPI assembles a JSONAPI DTO
func (org *Organization) ToJSONAPI() any {
	return &jsonapiOrganization{
		Name:            org.Name(),
		CreatedAt:       org.CreatedAt(),
		ExternalID:      org.ID(),
		Permissions:     &defaultOrganizationPermissions,
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	}
}

// organizationList represents a list of Organizations.
type organizationList struct {
	*otf.Pagination
	Items []*Organization
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *organizationList) ToJSONAPI() any {
	obj := &jsonapiList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, item.ToJSONAPI().(*jsonapiOrganization))
	}
	return obj
}

// OrganizationListOptions represents the options for listing organizations.
type OrganizationListOptions struct {
	otf.ListOptions
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
	if !otf.ValidStringID(opts.Name) {
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
