package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
// TFE.
type VCSProvider struct {
	id          string
	createdAt   time.Time
	name        string      // TODO: rename to description (?)
	cloudConfig CloudConfig // cloud config for creating client
	token       string      // credential for creating client

	organizationName string // vcs provider belongs to an organization
}

func (t *VCSProvider) ID() string               { return t.id }
func (t *VCSProvider) String() string           { return t.name }
func (t *VCSProvider) Token() string            { return t.token }
func (t *VCSProvider) CreatedAt() time.Time     { return t.createdAt }
func (t *VCSProvider) Name() string             { return t.name }
func (t *VCSProvider) OrganizationName() string { return t.organizationName }
func (t *VCSProvider) CloudConfig() CloudConfig { return t.cloudConfig }

func (t *VCSProvider) NewClient(ctx context.Context) (CloudClient, error) {
	return t.cloudConfig.NewClient(ctx, CloudCredentials{
		PersonalToken: &t.token,
	})
}

// VCSProviderFactory makes vcs providers
type VCSProviderFactory struct {
	CloudService
}

func (f *VCSProviderFactory) NewVCSProvider(opts VCSProviderCreateOptions) (*VCSProvider, error) {
	cloudConfig, err := f.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		name:             opts.Name,
		organizationName: opts.OrganizationName,
		cloudConfig:      cloudConfig,
		token:            opts.Token,
	}, nil
}

type VCSProviderCreateOptions struct {
	OrganizationName string
	Token            string
	Name             string
	Cloud            string
}

// VCSProviderRow represents a database row for a vcs provider
type VCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	Cloud            pgtype.Text        `json:"cloud"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// UnmarshalVCSProviderRow unmarshals a vcs provider row from the database.
func (u *Unmarshaler) UnmarshalVCSProviderRow(row VCSProviderRow) (*VCSProvider, error) {
	cloudConfig, err := u.GetCloudConfig(row.Cloud.String)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", cloudConfig)
	}

	return &VCSProvider{
		id:               row.VCSProviderID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		name:             row.Name.String,
		organizationName: row.OrganizationName.String,
		cloudConfig:      cloudConfig,
		token:            row.Token.String,
	}, nil
}

// VCSProviderService provides access to vcs providers
type VCSProviderService interface {
	CreateVCSProvider(ctx context.Context, opts VCSProviderCreateOptions) (*VCSProvider, error)
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id, organization string) error

	SetStatus(ctx context.Context, providerID string, opts SetStatusOptions) error
	GetRepository(ctx context.Context, providerID string, identifier string) (*Repo, error)
	GetRepoTarball(ctx context.Context, providerID string, opts GetRepoTarballOptions) ([]byte, error)
	ListRepositories(ctx context.Context, providerID string, opts ListOptions) (*RepoList, error)
	ListTags(ctx context.Context, providerID string, opts ListTagsOptions) ([]VCSRef, error)

	CreateWebhook(ctx context.Context, providerID string, opts CreateWebhookOptions) (string, error)
	UpdateWebhook(ctx context.Context, providerID string, opts UpdateWebhookOptions) error
	GetWebhook(ctx context.Context, providerID string, opts GetWebhookOptions) (*VCSWebhook, error)
	DeleteWebhook(ctx context.Context, providerID string, opts DeleteWebhookOptions) error
}

// VCSProviderStore persists vcs providers
type VCSProviderStore interface {
	CreateVCSProvider(ctx context.Context, provider *VCSProvider) error
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id string) error
}
