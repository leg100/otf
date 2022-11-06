package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

type VCSProvider struct {
	// TODO: do we need an id if name is unique?
	id        string
	createdAt time.Time
	token     string
	name      string
	cloud     string
	// vcs provider belongs to an organization
	organizationName string
}

func NewVCSProvider(opts VCSProviderCreateOptions) *VCSProvider {
	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		token:            opts.Token,
		name:             opts.Name,
		cloud:            opts.Cloud,
		organizationName: opts.OrganizationName,
	}
}

func (t *VCSProvider) ID() string               { return t.id }
func (t *VCSProvider) String() string           { return t.name }
func (t *VCSProvider) Token() string            { return t.token }
func (t *VCSProvider) CreatedAt() time.Time     { return t.createdAt }
func (t *VCSProvider) Name() string             { return t.name }
func (t *VCSProvider) Cloud() string            { return t.cloud }
func (t *VCSProvider) OrganizationName() string { return t.organizationName }

func (t *VCSProvider) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	switch t.cloud {
	case "github":
		return (&GithubCloud{defaultGithubConfig()}).NewDirectoryClient(ctx, opts)
	case "gitlab":
		return (&gitlabCloud{defaultGitlabConfig()}).NewDirectoryClient(ctx, opts)
	default:
		return nil, fmt.Errorf("vcs provider has no cloud specified")
	}
}

type VCSProviderCreateOptions struct {
	OrganizationName string `schema:"organization_name,required"`
	Token            string `schema:"token,required"`
	Name             string `schema:"name,required"`
	Cloud            string `schema:"cloud,required"`
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
func UnmarshalVCSProviderRow(row VCSProviderRow) *VCSProvider {
	return &VCSProvider{
		id:               row.VCSProviderID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		token:            row.Token.String,
		name:             row.Name.String,
		cloud:            row.Cloud.String,
		organizationName: row.OrganizationName.String,
	}
}

// VCSProviderService provides access to vcs providers
type VCSProviderService interface {
	CreateVCSProvider(ctx context.Context, opts VCSProviderCreateOptions) (*VCSProvider, error)
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
