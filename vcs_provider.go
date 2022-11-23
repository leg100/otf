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
	config ClientConfig

	id        string
	createdAt time.Time
	// TODO: name or description?
	name      string
	cloud     Cloud
	cloudName CloudName

	// vcs provider belongs to an organization
	organizationName string
}

func NewVCSProvider(opts VCSProviderCreateOptions) *VCSProvider {
	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		name:             opts.Name,
		organizationName: opts.OrganizationName,
		cloud:            opts.Cloud,
		cloudName:        opts.CloudName,
		config: ClientConfig{
			Hostname:            opts.Hostname,
			SkipTLSVerification: opts.SkipTLSVerification,
			PersonalToken:       String(opts.Token),
		},
	}
}

func (t *VCSProvider) ID() string                { return t.id }
func (t *VCSProvider) String() string            { return t.name }
func (t *VCSProvider) Token() string             { return *t.config.PersonalToken }
func (t *VCSProvider) Hostname() string          { return t.config.Hostname }
func (t *VCSProvider) CloudName() CloudName      { return t.cloudName }
func (t *VCSProvider) SkipTLSVerification() bool { return t.config.SkipTLSVerification }
func (t *VCSProvider) CreatedAt() time.Time      { return t.createdAt }
func (t *VCSProvider) Name() string              { return t.name }
func (t *VCSProvider) OrganizationName() string  { return t.organizationName }

func (t *VCSProvider) NewClient(ctx context.Context) (CloudClient, error) {
	return t.cloud.NewClient(ctx, t.config)
}

type VCSProviderCreateOptions struct {
	OrganizationName    string
	Token               string
	Name                string
	CloudName           CloudName
	Cloud               Cloud
	Hostname            string
	SkipTLSVerification bool
}

// VCSProviderRow represents a database row for a vcs provider
type VCSProviderRow struct {
	VCSProviderID       pgtype.Text        `json:"id"`
	Token               pgtype.Text        `json:"token"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	Name                pgtype.Text        `json:"name"`
	Hostname            pgtype.Text        `json:"hostname"`
	SkipTLSVerification bool               `json:"skip_tls_verification"`
	Cloud               pgtype.Text        `json:"cloud"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
}

// UnmarshalVCSProviderRow unmarshals a vcs provider row from the database.
func UnmarshalVCSProviderRow(row VCSProviderRow) (*VCSProvider, error) {
	var cloud Cloud
	switch CloudName(row.Cloud.String) {
	case GithubCloudName:
		cloud = GithubCloud{}
	case GitlabCloudName:
		cloud = GitlabCloud{}
	default:
		return nil, fmt.Errorf("unknown cloud: %s", cloud)
	}

	return &VCSProvider{
		id:               row.VCSProviderID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		name:             row.Name.String,
		organizationName: row.OrganizationName.String,
		cloud:            cloud,
		cloudName:        CloudName(row.Cloud.String),
		config: ClientConfig{
			Hostname:            row.Hostname.String,
			SkipTLSVerification: row.SkipTLSVerification,
			PersonalToken:       String(row.Token.String),
		},
	}, nil
}

// VCSProviderService provides access to vcs providers
type VCSProviderService interface {
	CreateVCSProvider(ctx context.Context, opts VCSProviderCreateOptions) (*VCSProvider, error)
	GetVCSProvider(ctx context.Context, id, organization string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id, organization string) error

	SetStatus(ctx context.Context, providerID string, opts SetStatusOptions) error
	GetRepository(ctx context.Context, providerID string, identifier string) (*Repo, error)
	GetRepoTarball(ctx context.Context, providerID string, opts GetRepoTarballOptions) ([]byte, error)
	ListRepositories(ctx context.Context, providerID string, opts ListOptions) (*RepoList, error)
	CreateWebhook(ctx context.Context, providerID string, opts CreateCloudWebhookOptions) error
	DeleteWebhook(ctx context.Context, providerID string, hook *Webhook) error
}

// VCSProviderStore persists vcs providers
type VCSProviderStore interface {
	CreateVCSProvider(ctx context.Context, provider *VCSProvider) error
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id string) error
}
