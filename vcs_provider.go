package otf

import (
	"context"
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

	*GithubCloud
}

func NewVCSProvider(opts VCSProviderCreateOptions) *VCSProvider {
	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		token:            opts.Token,
		name:             opts.Name,
		organizationName: opts.OrganizationName,
		GithubCloud:      &GithubCloud{defaultGithubConfig()},
	}
}

func (t *VCSProvider) ID() string               { return t.id }
func (t *VCSProvider) String() string           { return t.name }
func (t *VCSProvider) Token() string            { return t.token }
func (t *VCSProvider) CreatedAt() time.Time     { return t.createdAt }
func (t *VCSProvider) Name() string             { return t.name }
func (t *VCSProvider) OrganizationName() string { return t.organizationName }

type VCSProviderCreateOptions struct {
	OrganizationName string `schema:"organization_name,required"`
	Token            string `schema:"token,required"`
	Name             string `schema:"name,required"`
}

// VCSProviderRow represents a database row for a vcs provider
type VCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// UnmarshalVCSProviderRow unmarshals a vcs provider row from the database.
func UnmarshalVCSProviderRow(row VCSProviderRow) *VCSProvider {
	return &VCSProvider{
		id:               row.VCSProviderID.String,
		createdAt:        row.CreatedAt.Time.UTC(),
		token:            row.Token.String,
		name:             row.Name.String,
		organizationName: row.OrganizationName.String,
		GithubCloud:      &GithubCloud{defaultGithubConfig()},
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
