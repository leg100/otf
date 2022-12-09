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

func (m *Module) LatestVersion() *ModuleVersion {
	if len(m.versions) == 0 {
		return nil
	}
	sort.Sort(ByModuleVersion(m.versions))
	return m.versions[0]
}

type ModuleVersion struct {
	version   string
	createdAt time.Time
	updatedAt time.Time
	// TODO: download counter
}

type ModuleRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
}

type ModuleService interface {
	// TODO: rename option structs for *all* service endpoints, so that the name
	// reflects the method name, e.g. CreateModule -> CreateModuleOptions

	CreateModule(context.Context, ModuleCreateOptions) (*Module, error)
	CreateModuleVersion(context.Context, ModuleCreateVersionOptions) (*ModuleVersion, error)
	ListModules(context.Context, ModuleListOptions) ([]*Module, error)
	GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
	UploadModule(ctx context.Context, opts UploadModuleOptions) error
	DownloadModule(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
}

type (
	ModuleCreateOptions struct {
		Name         string
		Provider     string
		Repo         *ModuleRepo
		Organization *Organization
	}
	ModuleCreateVersionOptions struct {
		ModuleID string
		Version  string
	}
	GetModuleOptions struct {
		Name         string
		Provider     string
		Organization *Organization
	}
	UploadModuleOptions struct {
		Name     string
		Provider string
		Version  string
		Tarball  []byte
	}
	DownloadModuleOptions struct {
		Name     string
		Provider string
		Version  string
	}
	ModuleListOptions struct {
		Organization string // filter by organization name
	}
	ModuleList struct {
		*Pagination
		Items []*Module
	}
)

type ModuleMaker struct {
	ModuleService
	VCSProviderService
}

func (mm *ModuleMaker) NewModule(ctx context.Context, opts ModuleCreateOptions) (*Module, error) {
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

			modVersion, err := mm.CreateModuleVersion(ctx, ModuleCreateVersionOptions{
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
			err = mm.UploadModule(ctx, UploadModuleOptions{
				Tarball: tarball,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return mod, nil
}

func NewModule(opts ModuleCreateOptions) *Module {
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
