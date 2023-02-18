package organization

import (
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

func NewOrganization(opts otf.OrganizationCreateOptions) (*Organization, error) {
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

func (org *Organization) Update(opts UpdateOptions) error {
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
	return &JSONAPIOrganization{
		Name:            org.Name(),
		CreatedAt:       org.CreatedAt(),
		ExternalID:      org.ID(),
		Permissions:     &defaultOrganizationPermissions,
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	}
}

// OrganizationList represents a list of Organizations.
type OrganizationList struct {
	*otf.Pagination
	Items []*Organization
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *OrganizationList) ToJSONAPI() any {
	obj := &jsonapiList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, item.ToJSONAPI().(*JSONAPIOrganization))
	}
	return obj
}

// ListOptions represents the options for listing organizations.
type ListOptions struct {
	otf.ListOptions
}

// UpdateOptions represents the options for updating an
// organization.
type UpdateOptions struct {
	Name            *string
	SessionRemember *int
	SessionTimeout  *int
}
