package user

import (
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
)

type (
	// TFEUser represents an OTF user.
	TFEUser struct {
		ID               resource.TfeID `jsonapi:"primary,users"`
		AvatarURL        string         `jsonapi:"attribute" json:"avatar-url"`
		Email            string         `jsonapi:"attribute" json:"email"`
		IsServiceAccount bool           `jsonapi:"attribute" json:"is-service-account"`
		TwoFactor        *TFETwoFactor  `jsonapi:"attribute" json:"two-factor"`
		UnconfirmedEmail string         `jsonapi:"attribute" json:"unconfirmed-email"`
		Username         string         `jsonapi:"attribute" json:"username"`
		V2Only           bool           `jsonapi:"attribute" json:"v2-only"`

		// Relations
		// AuthenticationTokens *AuthenticationTokens `jsonapi:"relation,authentication-tokens"`
	}

	// TFETwoFactor represents the organization permissions.
	TFETwoFactor struct {
		Enabled  bool `jsonapi:"attribute" json:"enabled"`
		Verified bool `jsonapi:"attribute" json:"verified"`
	}

	// CreateUserOptions represents the options for creating a
	// user.
	TFECreateUserOptions struct {
		// Type is a public field utilized by JSON:API to
		// set the resource type via the field tag.
		// It is not a user-defined value and does not need to be set.
		// https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,users"`

		Username *string `jsonapi:"attribute" json:"username"`
	}
)

// TFEOrganizationMembership represents a Terraform Enterprise organization membership.
type TFEOrganizationMembership struct {
	ID     resource.TfeID                  `jsonapi:"primary,organization-memberships"`
	Status TFEOrganizationMembershipStatus `jsonapi:"attribute" json:"status"`
	Email  string                          `jsonapi:"attribute" json:"email"`

	// Relations
	Organization *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	User         *TFEUser                      `jsonapi:"relationship" json:"user"`
	Teams        []*team.TFETeam               `jsonapi:"relationship" json:"teams"`
}

// TFEOrganizationMembershipStatus represents an organization membership status.
type TFEOrganizationMembershipStatus string

const (
	OrganizationMembershipActive  TFEOrganizationMembershipStatus = "active"
	OrganizationMembershipInvited TFEOrganizationMembershipStatus = "invited"
)

// TFEOrganizationMembershipCreateOptions represents the options for creating an organization membership.
type TFEOrganizationMembershipCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organization-memberships"`

	// Required: User's email address.
	Email *string `jsonapi:"attribute" json:"email"`
}
