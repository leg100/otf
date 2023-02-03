// Package module is reponsible for registry modules
package module

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/semver"
	"github.com/leg100/otf/sql/pggen"
)

type ModuleStatus string

const (
	ModuleStatusPending       ModuleStatus = "pending"
	ModuleStatusNoVersionTags ModuleStatus = "no_version_tags"
	ModuleStatusSetupFailed   ModuleStatus = "setup_failed"
	ModuleStatusSetupComplete ModuleStatus = "setup_complete"
)

type Module struct {
	id           string
	createdAt    time.Time
	updatedAt    time.Time
	name         string
	provider     string
	organization string      // Module belongs to an organization
	repo         *ModuleRepo // Module optionally connected to vcs repo
	status       ModuleStatus
	versions     SortedModuleVersions
	latest       *ModuleVersion
}

func NewModule(opts CreateModuleOptions) *Module {
	return &Module{
		id:           otf.NewID("mod"),
		createdAt:    otf.CurrentTimestamp(),
		updatedAt:    otf.CurrentTimestamp(),
		name:         opts.Name,
		provider:     opts.Provider,
		status:       ModuleStatusPending,
		organization: opts.Organization,
		repo:         opts.Repo,
	}
}

func (m *Module) ID() string                     { return m.id }
func (m *Module) CreatedAt() time.Time           { return m.createdAt }
func (m *Module) UpdatedAt() time.Time           { return m.updatedAt }
func (m *Module) Name() string                   { return m.name }
func (m *Module) Provider() string               { return m.provider }
func (m *Module) Repo() *ModuleRepo              { return m.repo }
func (m *Module) Status() ModuleStatus           { return m.status }
func (m *Module) Versions() SortedModuleVersions { return m.versions }
func (m *Module) Latest() *ModuleVersion         { return m.versions.latest() }
func (m *Module) Organization() string           { return m.organization }

func (m *Module) UpdateStatus(status ModuleStatus) { m.status = status }

// addNewVersion creates a new module version and adds it to the module
func (m *Module) addNewVersion(opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	if m.Latest() != nil {
		// can only add a version greater than current version
		if cmp := semver.Compare(m.Latest().version, opts.Version); cmp >= 0 {
			return nil, errors.New("can only add version newer than current latest version")
		}
	}

	version := &ModuleVersion{
		id:        otf.NewID("modver"),
		createdAt: otf.CurrentTimestamp(),
		updatedAt: otf.CurrentTimestamp(),
		moduleID:  opts.ModuleID,
		version:   opts.Version,
		status:    ModuleVersionStatusPending,
	}
	m.versions = m.versions.add(version)

	return version, nil
}

// upload tarball for the module version: the callback should
// return the tarball to upload. The status of the module and the module version
// are set to reflect the success or failure of the callback
func (m *Module) upload(version string, tarballGetter func() ([]byte, error)) error {
	var modVer *ModuleVersion
	for _, v := range m.versions {
		if v.version == version {
			modVer = v
			break
		}
	}
	if modVer == nil {
		return errors.New("version does not exist")
	}

	tarball, err := tarballGetter()
	if err != nil {
		modVer.status = ModuleVersionStatusRegIngressFailed
		modVer.statusError = err.Error()
		return err
	}

	_, err = UnmarshalTerraformModule(tarball)
	if err != nil {
		modVer.status = ModuleVersionStatusCloneFailed
		modVer.statusError = err.Error()
		return err
	}

	modVer.status = ModuleVersionStatusOk

	return nil
}

// setStatus updates the module status to reflect the status of the given
// version
func (m *Module) setStatus(status ModuleVersionStatus) {
	if m.status == ModuleStatusSetupComplete {
		return
	}
	if status == ModuleVersionStatusOk {
		m.status = ModuleStatusSetupComplete
		return nil
	}
	if ver.status == ModuleVersionStatusPending {
		return nil
	}
	m.status = ModuleStatusSetupFailed
}

// Version returns the specified module version. If the empty string, then the
// latest version is returned. If there is no matching version or no versions at
// all then nil is returned.
func (m *Module) Version(version string) *ModuleVersion {
	if version == "" {
		return m.versions.latest()
	}
	for _, v := range m.versions {
		if v.version == version {
			return v
		}
	}
	return nil
}

func (m *Module) MarshalLog() any {
	return struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Provider     string `json:"provider"`
	}{
		ID:           m.id,
		Organization: m.organization,
		Name:         m.name,
		Provider:     m.provider,
	}
}

type ModuleRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
}

type ModuleService interface {
	// PublishModule publishes a module from a VCS repository.
	PublishModule(context.Context, PublishModuleOptions) (*Module, error)
	// CreateModule creates a module without a connection to a VCS repository.
	CreateModule(context.Context, CreateModuleOptions) (*Module, error)
	UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) (*Module, error)
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	GetModuleByID(ctx context.Context, id string) (*Module, error)
	GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
	DeleteModule(ctx context.Context, id string) (*Module, error)
}

type ModuleStore interface {
	CreateModule(context.Context, *Module) error
	UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) error
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	GetModuleByID(ctx context.Context, id string) (*Module, error)
	GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
	DeleteModule(ctx context.Context, id string) error
}

type ModuleDeleter interface {
	Delete(ctx context.Context, moduleID string) error
}

type (
	PublishModuleOptions struct {
		Identifier   string
		ProviderID   string
		Organization *otf.Organization
	}
	CreateModuleOptions struct {
		Name         string
		Provider     string
		Organization string
		Repo         *ModuleRepo
	}
	UpdateModuleStatusOptions struct {
		ID     string
		Status ModuleStatus
	}
	CreateModuleVersionOptions struct {
		ModuleID string
		Version  string
	}
	UpdateModuleVersionStatusOptions struct {
		ID     string
		Status ModuleVersionStatus
		Error  string
	}
	GetModuleOptions struct {
		Name         string
		Provider     string
		Organization string
	}
	UploadModuleVersionOptions struct {
		ModuleID      string
		Version       string
		TarballGetter func() ([]byte, error)
	}
	DownloadModuleOptions struct {
		ModuleVersionID string
	}
	ListModulesOptions struct {
		Organization string `schema:"organization_name,required"` // filter by organization name
	}
	ModuleList struct {
		*otf.Pagination
		Items []*Module
	}
)

// SortedModuleVersions is a list of module versions belonging to module, sorted
// by their semantic version, oldest version first
type SortedModuleVersions []*ModuleVersion

// add adds a module version and returns the new list in sorted order
func (l SortedModuleVersions) add(modver *ModuleVersion) SortedModuleVersions {
	newl := append(l, modver)
	sort.Sort(newl)
	return newl
}

// LatestVersion returns the latest ok version
func (l SortedModuleVersions) latest() *ModuleVersion {
	// starting from the
	for i := len(l) - 1; i >= 0; i-- {
		if l[i].Status() == ModuleVersionStatusOk {
			return l[i]
		}
	}
	return nil
}

func (l SortedModuleVersions) Len() int { return len(l) }
func (l SortedModuleVersions) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l SortedModuleVersions) Less(i, j int) bool {
	// TODO: sort out this suspect logic
	cmp := semver.Compare(l[i].version, l[j].version)
	if cmp != 0 {
		return cmp < 0
	}
	return l[i].version < l[j].version
}

// ModuleRow is a row from a database query for modules.
type ModuleRow struct {
	ModuleID         pgtype.Text            `json:"module_id"`
	CreatedAt        pgtype.Timestamptz     `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz     `json:"updated_at"`
	Name             pgtype.Text            `json:"name"`
	Provider         pgtype.Text            `json:"provider"`
	Status           pgtype.Text            `json:"status"`
	OrganizationName pgtype.Text            `json:"organization_name"`
	ModuleRepo       *pggen.ModuleRepos     `json:"module_repo"`
	Webhook          *pggen.Webhooks        `json:"webhook"`
	Versions         []pggen.ModuleVersions `json:"versions"`
}

// UnmarshalModuleRow unmarshals a database row into a module
func UnmarshalModuleRow(row ModuleRow) *Module {
	module := &Module{
		id:           row.ModuleID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		updatedAt:    row.UpdatedAt.Time.UTC(),
		name:         row.Name.String,
		provider:     row.Provider.String,
		status:       ModuleStatus(row.Status.String),
		organization: row.OrganizationName.String,
	}
	if row.ModuleRepo != nil {
		module.repo = &ModuleRepo{
			ProviderID: row.ModuleRepo.VCSProviderID.String,
			WebhookID:  row.Webhook.WebhookID.Bytes,
			Identifier: row.Webhook.Identifier.String,
		}
	}
	for _, version := range row.Versions {
		module.Add(UnmarshalModuleVersionRow(ModuleVersionRow(version)))
	}
	return module
}

// ListModuleRepositories wraps the ListRepositories endpoint, returning only
// those repositories with a name matching the format
// '<something>-<provider>-<module>'.
//
// NOTE: no pagination is performed, only matching results from the first page
// are retrieved
func ListModuleRepositories(ctx context.Context, app otf.Application, providerID string) ([]cloud.Repo, error) {
	client, err := app.GetVCSClient(ctx, providerID)
	if err != nil {
		return nil, err
	}

	list, err := client.ListRepositories(ctx, cloud.ListRepositoriesOptions{
		PageSize: otf.MaxPageSize,
	})
	if err != nil {
		return nil, err
	}
	var filtered []cloud.Repo
	for _, repo := range list {
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
