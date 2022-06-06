package otf

import (
	"context"
	"net/http"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
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

// ToJSONAPI assembles a JSONAPI DTO
func (org *Organization) ToJSONAPI(req *http.Request) any {
	return &jsonapi.Organization{
		Name:            org.Name(),
		CreatedAt:       org.CreatedAt(),
		ExternalID:      org.ID(),
		Permissions:     &jsonapi.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	}
}

// OrganizationList represents a list of Organizations.
type OrganizationList struct {
	*Pagination
	Items []*Organization
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *OrganizationList) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.OrganizationList{
		Pagination: (*jsonapi.Pagination)(l.Pagination),
	}
	for _, item := range l.Items {
		dto.Items = append(dto.Items, item.ToJSONAPI(req).(*jsonapi.Organization))
	}
	return dto
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
	Create(ctx context.Context, org *Organization) error
	Get(ctx context.Context, name string) (*Organization, error)
	List(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
	Update(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error)
	Delete(ctx context.Context, name string) error
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
