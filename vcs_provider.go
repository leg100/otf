package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

const (
	VCSProviderGithub VCSProviderCloud = "github"
	VCSProviderGitlab VCSProviderCloud = "gitlab"
)

type VCSProviderCloud string

// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
// TFE.
type VCSProvider struct {
	// TODO: do we need an id if name is unique?
	id        string
	createdAt time.Time
	token     string
	name      string
	cloud     Cloud
	// vcs provider belongs to an organization
	organizationName string
}

func NewVCSProvider(cloud Cloud, opts VCSProviderCreateOptions) (*VCSProvider, error) {
	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		token:            opts.Token,
		name:             opts.Name,
		cloud:            cloud,
		organizationName: opts.OrganizationName,
	}, nil
}

func (t *VCSProvider) ID() string               { return t.id }
func (t *VCSProvider) String() string           { return t.name }
func (t *VCSProvider) Token() string            { return t.token }
func (t *VCSProvider) CreatedAt() time.Time     { return t.createdAt }
func (t *VCSProvider) Name() string             { return t.name }
func (t *VCSProvider) Cloud() Cloud             { return t.cloud }
func (t *VCSProvider) OrganizationName() string { return t.organizationName }

func (t *VCSProvider) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	return t.cloud.NewDirectoryClient(ctx, opts)
}

type VCSProviderCreateOptions struct {
	OrganizationName string `schema:"organization_name,required"`
	Token            string `schema:"token,required"`
	Name             string `schema:"name,required"`
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
	provider := &VCSProvider{
		id:               row.VCSProviderID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		token:            row.Token.String,
		name:             row.Name.String,
		organizationName: row.OrganizationName.String,
	}

	// unmarshal provider cloud
	opts := cloudConfigOptions{
		hostname:            String(row.Hostname.String),
		skipTLSVerification: Bool(row.SkipTLSVerification),
	}
	switch row.Cloud.String {
	case "github":
		provider.cloud = NewGithubCloud(&opts)
	case "gitlab":
		provider.cloud = NewGitlabCloud(&opts)
	default:
		return nil, fmt.Errorf("unknown cloud: %s", row.Cloud.String)
	}

	return provider, nil
}

// VCSProviderService provides access to vcs providers
type VCSProviderService interface {
	CreateVCSProvider(ctx context.Context, cloud Cloud, opts VCSProviderCreateOptions) (*VCSProvider, error)
	GetVCSProvider(ctx context.Context, id, organization string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id, organization string) error
}

// VCSProviderStore persists vcs providers
type VCSProviderStore interface {
	CreateVCSProvider(ctx context.Context, provider *VCSProvider) error
	GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
	ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
	DeleteVCSProvider(ctx context.Context, id string) error
}
