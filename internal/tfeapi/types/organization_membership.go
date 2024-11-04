package types

import "github.com/leg100/otf/internal/resource"

// OrganizationMembership represents a Terraform Enterprise organization membership.
type OrganizationMembership struct {
	ID     resource.ID                  `jsonapi:"primary,organization-memberships"`
	Status OrganizationMembershipStatus `jsonapi:"attribute" json:"status"`
	Email  string                       `jsonapi:"attribute" json:"email"`

	// Relations
	Organization *Organization `jsonapi:"relationship" json:"organization"`
	User         *User         `jsonapi:"relationship" json:"user"`
	Teams        []*Team       `jsonapi:"relationship" json:"teams"`
}

// OrganizationMembershipStatus represents an organization membership status.
type OrganizationMembershipStatus string

const (
	OrganizationMembershipActive  OrganizationMembershipStatus = "active"
	OrganizationMembershipInvited OrganizationMembershipStatus = "invited"
)

// OrganizationMembershipCreateOptions represents the options for creating an organization membership.
type OrganizationMembershipCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organization-memberships"`

	// Required: User's email address.
	Email *string `jsonapi:"attribute" json:"email"`
}
