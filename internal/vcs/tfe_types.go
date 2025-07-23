package vcs

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

// TFEServiceProviderType represents a VCS type.
type TFEServiceProviderType string

// List of available VCS types.
const (
	ServiceProviderAzureDevOpsServer   TFEServiceProviderType = "ado_server"
	ServiceProviderAzureDevOpsServices TFEServiceProviderType = "ado_services"
	ServiceProviderBitbucket           TFEServiceProviderType = "bitbucket_hosted"
	// Bitbucket Server v5.4.0 and above
	ServiceProviderBitbucketServer TFEServiceProviderType = "bitbucket_server"
	// Bitbucket Server v5.3.0 and below
	ServiceProviderBitbucketServerLegacy TFEServiceProviderType = "bitbucket_server_legacy"
	ServiceProviderForgejo               TFEServiceProviderType = "forgejo"
	ServiceProviderGithub                TFEServiceProviderType = "github"
	ServiceProviderGithubEE              TFEServiceProviderType = "github_enterprise"
	ServiceProviderGithubApp             TFEServiceProviderType = "github_app"
	ServiceProviderGitlab                TFEServiceProviderType = "gitlab_hosted"
	ServiceProviderGitlabCE              TFEServiceProviderType = "gitlab_community_edition"
	ServiceProviderGitlabEE              TFEServiceProviderType = "gitlab_enterprise_edition"
)

// TFEOAuthClient represents a connection between an organization and a VCS
// provider.
type TFEOAuthClient struct {
	ID                  resource.TfeID         `jsonapi:"primary,oauth-clients"`
	APIURL              string                 `jsonapi:"attribute" json:"api-url"`
	CallbackURL         string                 `jsonapi:"attribute" json:"callback-url"`
	ConnectPath         string                 `jsonapi:"attribute" json:"connect-path"`
	CreatedAt           time.Time              `jsonapi:"attribute" json:"created-at"`
	HTTPURL             string                 `jsonapi:"attribute" json:"http-url"`
	Key                 string                 `jsonapi:"attribute" json:"key"`
	RSAPublicKey        string                 `jsonapi:"attribute" json:"rsa-public-key"`
	Name                *string                `jsonapi:"attribute" json:"name"`
	Secret              string                 `jsonapi:"attribute" json:"secret"`
	ServiceProvider     TFEServiceProviderType `jsonapi:"attribute" json:"service-provider"`
	ServiceProviderName string                 `jsonapi:"attribute" json:"service-provider-display-name"`

	// Relations
	Organization *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	OAuthTokens  []*TFEOAuthToken              `jsonapi:"relationship" json:"oauth-tokens"`
}

// TFEOAuthClientCreateOptions represents the options for creating an OAuth client.
type TFEOAuthClientCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,oauth-clients"`

	// A display name for the OAuth Client.
	Name *string `jsonapi:"attribute" json:"name"`

	// Required: The base URL of your VCS provider's API.
	APIURL *internal.WebURL `jsonapi:"attribute" json:"api-url"`

	// Required: The homepage of your VCS provider.
	HTTPURL *internal.WebURL `jsonapi:"attribute" json:"http-url"`

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
	ServiceProvider *TFEServiceProviderType `jsonapi:"attribute" json:"service-provider"`
}

// TFEOAuthToken represents a VCS configuration including the associated
// OAuth token
type TFEOAuthToken struct {
	ID                  resource.TfeID `jsonapi:"primary,oauth-tokens"`
	UID                 resource.TfeID `jsonapi:"attribute" json:"uid"`
	CreatedAt           time.Time      `jsonapi:"attribute" json:"created-at"`
	HasSSHKey           bool           `jsonapi:"attribute" json:"has-ssh-key"`
	ServiceProviderUser string         `jsonapi:"attribute" json:"service-provider-user"`

	// Relations
	OAuthClient *TFEOAuthClient `jsonapi:"relationship" json:"oauth-client"`
}
