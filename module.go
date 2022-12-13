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
	return m.versions[0]
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

type ModuleVersion struct {
	id        string
	moduleID  string
	version   string
	createdAt time.Time
	updatedAt time.Time
	// TODO: download counter
}

func (v *ModuleVersion) ID() string           { return v.id }
func (v *ModuleVersion) ModuleID() string     { return v.moduleID }
func (v *ModuleVersion) Version() string      { return v.version }
func (v *ModuleVersion) CreatedAt() time.Time { return v.createdAt }
func (v *ModuleVersion) UpdatedAt() time.Time { return v.updatedAt }

func (v *ModuleVersion) MarshalLog() any {
	return struct {
		ID, ModuleID, Version string
	}{
		ID:       v.id,
		ModuleID: v.moduleID,
		Version:  v.version,
	}
}

type ModuleRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
}

type ModuleService interface {
	CreateModule(context.Context, CreateModuleOptions) (*Module, error)
	CreateModuleVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
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
	CreateModuleOptions struct {
		Name         string
		Provider     string
		Repo         *ModuleRepo
		Organization *Organization
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

// ModuleMaker makes new modules, including its versions, one for each tag found
// in its connected repo (if connected to a repo).
type ModuleMaker struct {
	Application
	*ModulePublisher
}

func (mm *ModuleMaker) NewModule(ctx context.Context, opts CreateModuleOptions) (*Module, error) {
	mod := NewModule(opts)

	// If not connected to a repo there is nothing more to be done.
	if opts.Repo == nil {
		return mod, nil
	}

	// Make new version for each tag that looks like a semantic version.
	tags, err := mm.ListTags(ctx, opts.Repo.ProviderID, ListTagsOptions{
		Identifier: opts.Repo.Identifier,
	})
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		_, version, found := strings.Cut(string(tag), "/")
		if !found {
			return nil, fmt.Errorf("malformed git ref: %s", tag)
		}

		// skip tags that are not semantic versions
		if !semver.IsValid(version) {
			continue
		}

		modVersion, err := mm.Publish(ctx, PublishModuleVersionOptions{
			ModuleID: mod.ID(),
			// strip off v prefix if it has one
			Version:    strings.TrimPrefix(version, "v"),
			Ref:        string(tag),
			Identifier: opts.Repo.Identifier,
			ProviderID: mod.Repo().ProviderID,
		})
		if err != nil {
			return nil, err
		}
		mod.versions = append(mod.versions, modVersion)
	}
	return mod, nil
}

func NewModule(opts CreateModuleOptions) *Module {
	m := Module{
		id:           NewID("mod"),
		createdAt:    CurrentTimestamp(),
		updatedAt:    CurrentTimestamp(),
		name:         opts.Name,
		provider:     opts.Provider,
		organization: opts.Organization,
	}
	return &m
}

func NewModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		id:        NewID("mv"),
		createdAt: CurrentTimestamp(),
		updatedAt: CurrentTimestamp(),
		moduleID:  opts.ModuleID,
		version:   opts.Version,
	}
}

// ModulePublisher publishes new module versions.
type ModulePublisher struct {
	Application
}

type PublishModuleVersionOptions struct {
	ModuleID   string
	Version    string
	Ref        string
	Identifier string
	ProviderID string
}

// Publish a module version in response to a vcs event.
func (p *ModulePublisher) PublishFromEvent(ctx context.Context, event VCSEvent) error {
	// only publish when new tag is created
	tag, ok := event.(*VCSTagEvent)
	if !ok {
		return nil
	}
	if tag.Action != VCSTagEventCreatedAction {
		return nil
	}
	// only interested in tags that look like semantic versions
	if !semver.IsValid(tag.Tag) {
		return nil
	}

	module, err := p.GetModuleByWebhookID(ctx, tag.WebhookID)
	if err != nil {
		return err
	}
	if module.Repo() == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.ID())
	}

	// skip older or equal versions
	currentVersion := module.LatestVersion().Version()
	if n := semver.Compare(tag.Tag, currentVersion); n <= 0 {
		return nil
	}

	_, err = p.Publish(ctx, PublishModuleVersionOptions{
		ModuleID: module.ID(),
		// strip off v prefix if it has one
		Version:    strings.TrimPrefix(tag.Tag, "v"),
		Ref:        tag.CommitSHA,
		Identifier: tag.Identifier,
		ProviderID: module.Repo().ProviderID,
	})
	if err != nil {
		return err
	}

	return nil
}

// Publish a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (p *ModulePublisher) Publish(ctx context.Context, opts PublishModuleVersionOptions) (*ModuleVersion, error) {
	version, err := p.CreateModuleVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return nil, err
	}

	tarball, err := p.GetRepoTarball(ctx, opts.ProviderID, GetRepoTarballOptions{
		Identifier: opts.Identifier,
		Ref:        opts.Ref,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tarball: %w", err)
	}

	// upload tarball
	err = p.UploadModuleVersion(ctx, UploadModuleVersionOptions{
		ModuleVersionID: version.ID(),
		Tarball:         tarball,
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}

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

// UnmarshalModuleRow unmarshals a module database row into a module
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
