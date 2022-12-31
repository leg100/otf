package otf

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/semver"
	"github.com/leg100/otf/sql/pggen"
)

type Module struct {
	id           string
	createdAt    time.Time
	updatedAt    time.Time
	name         string
	provider     string
	organization *Organization // Module belongs to an organization
	repo         *ModuleRepo   // Module optionally connected to vcs repo
	versions     []*ModuleVersion
}

func NewModule(opts CreateModuleOptions) *Module {
	return &Module{
		id:           NewID("mod"),
		createdAt:    CurrentTimestamp(),
		updatedAt:    CurrentTimestamp(),
		name:         opts.Name,
		provider:     opts.Provider,
		organization: opts.Organization,
		repo:         opts.Repo,
	}
}

func (m *Module) ID() string                  { return m.id }
func (m *Module) CreatedAt() time.Time        { return m.createdAt }
func (m *Module) UpdatedAt() time.Time        { return m.updatedAt }
func (m *Module) Name() string                { return m.name }
func (m *Module) Provider() string            { return m.provider }
func (m *Module) Repo() *ModuleRepo           { return m.repo }
func (m *Module) Versions() []*ModuleVersion  { return m.versions }
func (m *Module) Organization() *Organization { return m.organization }

func (m *Module) LatestVersion() *ModuleVersion {
	if len(m.versions) == 0 {
		return nil
	}
	sort.Sort(ByModuleVersion(m.versions))
	return m.versions[len(m.versions)-1]
}

// Version returns the specified module version. If the empty string, then the
// latest version is returned. If there is no matching version or no versions at
// all then nil is returned.
func (m *Module) Version(version string) *ModuleVersion {
	if version == "" {
		return m.LatestVersion()
	}
	for _, v := range m.versions {
		if v.version == version {
			return v
		}
	}
	return nil
}

func (v *Module) MarshalLog() any {
	return struct {
		ID, Organization, Name, Provider string
	}{
		ID:           v.id,
		Organization: v.organization.name,
		Name:         v.name,
		Provider:     v.provider,
	}
}

type ModuleRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
}

type ModuleService interface {
	// PublishModule publishes a module from a VCS repository.
	PublishModule(context.Context, PublishModuleOptions) (*Module, error)
	// CreateModule creates a module without a connection to a VCS repository.
	CreateModule(context.Context, CreateModuleOptions) (*Module, error)
	CreateModuleVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	GetModuleByID(ctx context.Context, id string) (*Module, error)
	GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
	DeleteModule(ctx context.Context, id string) error
}

type ModuleStore interface {
	CreateModule(context.Context, *Module) error
	CreateModuleVersion(context.Context, *ModuleVersion) error
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	GetModuleByID(ctx context.Context, id string) (*Module, error)
	GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
	DeleteModule(ctx context.Context, id string) error
}

type (
	PublishModuleOptions struct {
		Identifier   string
		ProviderID   string
		OTFHost      string
		Organization *Organization
	}
	CreateModuleOptions struct {
		Name         string
		Provider     string
		Organization *Organization
		Repo         *ModuleRepo
	}
	CreateModuleVersionOptions struct {
		ModuleID string
		Version  string
	}
	GetModuleOptions struct {
		Name         string
		Provider     string
		Organization string
	}
	UploadModuleVersionOptions struct {
		ModuleVersionID string
		Tarball         []byte
	}
	DownloadModuleOptions struct {
		ModuleVersionID string
	}
	ListModulesOptions struct {
		Organization string `schema:"organization_name,required"` // filter by organization name
	}
	ModuleList struct {
		*Pagination
		Items []*Module
	}
)

// ByModuleVersion implements sort.Interface for sorting module versions.
type ByModuleVersion []*ModuleVersion

func (l ByModuleVersion) Len() int { return len(l) }
func (l ByModuleVersion) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l ByModuleVersion) Less(i, j int) bool {
	cmp := semver.Compare(l[i].version, l[j].version)
	if cmp != 0 {
		return cmp < 0
	}
	return l[i].version < l[j].version
}

// ModuleRow is a row from a database query for modules.
type ModuleRow struct {
	ModuleID     pgtype.Text            `json:"module_id"`
	CreatedAt    pgtype.Timestamptz     `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz     `json:"updated_at"`
	Name         pgtype.Text            `json:"name"`
	Provider     pgtype.Text            `json:"provider"`
	Organization *pggen.Organizations   `json:"organization"`
	ModuleRepo   *pggen.ModuleRepos     `json:"module_repo"`
	Webhook      *pggen.Webhooks        `json:"webhook"`
	Versions     []pggen.ModuleVersions `json:"versions"`
}

// UnmarshalModuleRow unmarshals a database row into a module
func UnmarshalModuleRow(row ModuleRow) *Module {
	module := &Module{
		id:           row.ModuleID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		updatedAt:    row.UpdatedAt.Time.UTC(),
		name:         row.Name.String,
		provider:     row.Provider.String,
		organization: UnmarshalOrganizationRow(*row.Organization),
	}
	if row.ModuleRepo != nil {
		module.repo = &ModuleRepo{
			ProviderID: row.ModuleRepo.VCSProviderID.String,
			WebhookID:  row.Webhook.WebhookID.Bytes,
			Identifier: row.Webhook.Identifier.String,
			HTTPURL:    row.Webhook.HTTPURL.String,
		}
	}
	for _, version := range row.Versions {
		module.versions = append(module.versions, &ModuleVersion{
			id:        version.ModuleVersionID.String,
			version:   version.Version.String,
			createdAt: row.CreatedAt.Time.UTC(),
			updatedAt: row.UpdatedAt.Time.UTC(),
			moduleID:  row.ModuleID.String,
		})
	}
	return module
}

// ListModuleRepositories wraps the ListRepositories endpoint, returning only
// those repositories with a name matching the format
// '<something>-<provider>-<module>'.
//
// NOTE: no pagination is performed, only matching results from the first page
// are retrieved
func ListModuleRepositories(ctx context.Context, app Application, providerID string) ([]*Repo, error) {
	list, err := app.ListRepositories(ctx, providerID, ListOptions{
		PageSize: MaxPageSize,
	})
	if err != nil {
		return nil, err
	}
	var filtered []*Repo
	for _, repo := range list.Items {
		_, name, found := strings.Cut(repo.Identifier, "/")
		if !found {
			return nil, fmt.Errorf("malformed identifier: %s", repo.Identifier)
		}
		parts := strings.SplitN(name, "-", 3)
		if len(parts) >= 3 {
			filtered = append(filtered, repo)
		}
	}
	return filtered, nil
}
