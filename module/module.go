// Package module is reponsible for registry modules
package module

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/semver"
)

const (
	ModuleStatusPending       ModuleStatus = "pending"
	ModuleStatusNoVersionTags ModuleStatus = "no_version_tags"
	ModuleStatusSetupFailed   ModuleStatus = "setup_failed"
	ModuleStatusSetupComplete ModuleStatus = "setup_complete"
)

type (
	ModuleStatus string

	Module struct {
		id           string
		createdAt    time.Time
		updatedAt    time.Time
		name         string
		provider     string
		organization string // Module belongs to an organization
		status       ModuleStatus
		versions     map[string]*ModuleVersion
		latest       *ModuleVersion

		repo *otf.Connection // optional vcs repo connection
	}

	ModuleDeleter interface {
		Delete(ctx context.Context, moduleID string) error
	}

	PublishModuleOptions struct {
		Identifier   string
		ProviderID   string
		Organization string
	}
	CreateModuleOptions struct {
		Name         string
		Provider     string
		Organization string
		Repo         *otf.Connection
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
		Name         string `schema:"name,required"`
		Provider     string `schema:"provider,required"`
		Organization string `schema:"organization,required"`
	}
	UploadModuleVersionOptions struct {
		VersionID string
		Tarball   []byte
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

	// SortedModuleVersions is a list of module versions belonging to module, sorted
	// by their semantic version, oldest version first
	SortedModuleVersions []*ModuleVersion
)

func newModule(opts CreateModuleOptions) *Module {
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

func (m *Module) UpdateStatus(status ModuleStatus) { m.status = status }

// addNewVersion creates a new module version and adds it to the module
func (m *Module) addNewVersion(opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	if m.latest != nil {
		// can only add a version greater than current version
		if cmp := semver.Compare(m.latest.version, opts.Version); cmp >= 0 {
			return nil, errors.New("can only add version newer than current latest version")
		}
	}

	modVer := &ModuleVersion{
		id:        otf.NewID("modver"),
		createdAt: otf.CurrentTimestamp(),
		updatedAt: otf.CurrentTimestamp(),
		moduleID:  opts.ModuleID,
		version:   opts.Version,
		status:    ModuleVersionStatusPending,
	}
	m.versions[modVer.version] = modVer

	return modVer, nil
}

// updateVersionStatus updates a module version's status - and error status if
// applicable.
func (m *Module) updateVersionStatus(version string, status ModuleVersionStatus, err error) error {
	modver, ok := m.versions[version]
	if !ok {
		return fmt.Errorf("version %s not found", version)
	}

	if err != nil {
		modver.statusError = err.Error()
	}
	modver.status = status
	m.setLatest()

	return nil
}

// setLatest sets the latest version, which is the greatest version with a
// healthy status. The module status is updated to reflect the latest version.
func (m *Module) setLatest() {
	versions := make([]string, len(m.versions))
	for v := range m.versions {
		versions = append(versions, v)
	}
	semver.Sort(versions)
	for i := len(versions) - 1; i >= 0; i-- {
		if m.versions[versions[i]].status == ModuleVersionStatusOk {
			m.latest = m.versions[versions[i]]
			m.status = ModuleStatus(ModuleVersionStatusOk)
			return
		}
	}
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

// add adds a module version and returns the new list in sorted order
func (l SortedModuleVersions) add(modver *ModuleVersion) SortedModuleVersions {
	newl := append(l, modver)
	sort.Sort(newl)
	return newl
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

// listTerraformModuleRepos wraps a cloud's ListRepositories endpoint, returning only
// those repositories with a name matching the format
// '<something>-<provider>-<module>'.
//
// NOTE: no pagination is performed, and only matching results from the first page
// are retrieved
func listTerraformModuleRepos(ctx context.Context, client cloud.Client) ([]cloud.Repo, error) {
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
