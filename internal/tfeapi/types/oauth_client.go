// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// ServiceProviderType represents a VCS type.
type ServiceProviderType string

// List of available VCS types.
const (
	ServiceProviderAzureDevOpsServer   ServiceProviderType = "ado_server"
	ServiceProviderAzureDevOpsServices ServiceProviderType = "ado_services"
	ServiceProviderBitbucket           ServiceProviderType = "bitbucket_hosted"
	// Bitbucket Server v5.4.0 and above
	ServiceProviderBitbucketServer ServiceProviderType = "bitbucket_server"
	// Bitbucket Server v5.3.0 and below
	ServiceProviderBitbucketServerLegacy ServiceProviderType = "bitbucket_server_legacy"
	ServiceProviderGithub                ServiceProviderType = "github"
	ServiceProviderGithubEE              ServiceProviderType = "github_enterprise"
	ServiceProviderGitlab                ServiceProviderType = "gitlab_hosted"
	ServiceProviderGitlabCE              ServiceProviderType = "gitlab_community_edition"
	ServiceProviderGitlabEE              ServiceProviderType = "gitlab_enterprise_edition"
)

// OAuthClient represents a connection between an organization and a VCS
// provider.
type OAuthClient struct {
	ID                  resource.TfeID      `jsonapi:"primary,oauth-clients"`
	APIURL              string              `jsonapi:"attribute" json:"api-url"`
	CallbackURL         string              `jsonapi:"attribute" json:"callback-url"`
	ConnectPath         string              `jsonapi:"attribute" json:"connect-path"`
	CreatedAt           time.Time           `jsonapi:"attribute" json:"created-at"`
	HTTPURL             string              `jsonapi:"attribute" json:"http-url"`
	Key                 string              `jsonapi:"attribute" json:"key"`
	RSAPublicKey        string              `jsonapi:"attribute" json:"rsa-public-key"`
	Name                *string             `jsonapi:"attribute" json:"name"`
	Secret              string              `jsonapi:"attribute" json:"secret"`
	ServiceProvider     ServiceProviderType `jsonapi:"attribute" json:"service-provider"`
	ServiceProviderName string              `jsonapi:"attribute" json:"service-provider-display-name"`

	// Relations
	Organization *Organization `jsonapi:"relationship" json:"organization"`
	OAuthTokens  []*OAuthToken `jsonapi:"relationship" json:"oauth-tokens"`
}

// OAuthClientCreateOptions represents the options for creating an OAuth client.
type OAuthClientCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,oauth-clients"`

	// A display name for the OAuth Client.
	Name *string `jsonapi:"attribute" json:"name"`

	// Required: The base URL of your VCS provider's API.
	APIURL *string `jsonapi:"attribute" json:"api-url"`

	// Required: The homepage of your VCS provider.
	HTTPURL *string `jsonapi:"attribute" json:"http-url"`

	// Optional: The OAuth Client key.
	Key *string `jsonapi:"attribute" json:"key,omitempty"`

	// Optional: The token string you were given by your VCS provider.
	OAuthToken *string `jsonapi:"attribute" json:"oauth-token-string,omitempty"`

	// Optional: Private key associated with this vcs provider - only available for ado_server
	PrivateKey *string `jsonapi:"attribute" json:"private-key,omitempty"`

	// Optional: Secret key associated with this vcs provider - only available for ado_server
	Secret *string `jsonapi:"attribute" json:"secret,omitempty"`

	// Optional: RSAPublicKey the text of the SSH public key associated with your BitBucket
	// Server Application Link.
	RSAPublicKey *string `jsonapi:"attribute" json:"rsa-public-key,omitempty"`

	// Required: The VCS provider being connected with.
	ServiceProvider *ServiceProviderType `jsonapi:"attribute" json:"service-provider"`
}
