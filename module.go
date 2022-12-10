package otf

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/mod/semver"
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

type ModuleVersion struct {
	moduleID  string
	version   string
	createdAt time.Time
	updatedAt time.Time
	// TODO: download counter
}

func (v *ModuleVersion) ModuleID() string     { return v.moduleID }
func (v *ModuleVersion) Version() string      { return v.version }
func (v *ModuleVersion) CreatedAt() time.Time { return v.createdAt }
func (v *ModuleVersion) UpdatedAt() time.Time { return v.updatedAt }

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
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
}

type ModuleStore interface {
	CreateModule(context.Context, *Module) error
	CreateModuleVersion(context.Context, *ModuleVersion) (*ModuleVersion, error)
	ListModules(context.Context, ListModulesOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
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
		Organization *Organization
	}
	UploadModuleVersionOptions struct {
		ModuleID string
		Version  string
		Tarball  []byte
	}
	DownloadModuleOptions struct {
		ModuleID string
		Version  string
	}
	ListModulesOptions struct {
		Organization string // filter by organization name
	}
	ModuleList struct {
		*Pagination
		Items []*Module
	}
)

type ModuleMaker struct {
	Application
}

func (mm *ModuleMaker) NewModule(ctx context.Context, opts CreateModuleOptions) (*Module, error) {
	mod := NewModule(opts)

	if opts.Repo != nil {
		// list all tags starting with 'v' in the module's repo
		tags, err := mm.ListTags(ctx, opts.Repo.ProviderID, ListTagsOptions{
			Identifier: opts.Repo.Identifier,
		})
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			_, version, found := strings.Cut(tag.Ref, "/")
			if !found {
				return nil, fmt.Errorf("malformed git tag ref: %s", tag.Ref)
			}

			// skip tags that are not semantic versions
			if !isValidSemVer(version) {
				continue
			}

			// strip off 'v' prefix if it has one
			version = strings.TrimPrefix(version, "v")

			modVersion, err := mm.CreateModuleVersion(ctx, CreateModuleVersionOptions{
				ModuleID: mod.id,
				Version:  version,
			})
			if err != nil {
				return nil, err
			}
			mod.versions = append(mod.versions, modVersion)

			tarball, err := mm.GetRepoTarball(ctx, opts.Repo.ProviderID, GetRepoTarballOptions{
				Ref: tag.Ref,
			})
			if err != nil {
				return nil, err
			}

			// upload tarball
			err = mm.UploadModuleVersion(ctx, UploadModuleVersionOptions{
				Tarball: tarball,
			})
			if err != nil {
				return nil, err
			}
		}
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

func isValidSemVer(s string) bool {
	// semver lib requires 'v' prefix
	if !strings.HasPrefix(s, "v") {
		s = "v" + s
	}
	return semver.IsValid(s)
}
