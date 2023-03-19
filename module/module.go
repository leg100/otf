// Package module is reponsible for registry modules
package module

import (
	"errors"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/repo"
)

const (
	ModuleStatusPending       ModuleStatus = "pending"
	ModuleStatusNoVersionTags ModuleStatus = "no_version_tags"
	ModuleStatusSetupFailed   ModuleStatus = "setup_failed"
	ModuleStatusSetupComplete ModuleStatus = "setup_complete"

	ModuleVersionStatusPending             ModuleVersionStatus = "pending"
	ModuleVersionStatusCloning             ModuleVersionStatus = "cloning"
	ModuleVersionStatusCloneFailed         ModuleVersionStatus = "clone_failed"
	ModuleVersionStatusRegIngressReqFailed ModuleVersionStatus = "reg_ingress_req_failed"
	ModuleVersionStatusRegIngressing       ModuleVersionStatus = "reg_ingressing"
	ModuleVersionStatusRegIngressFailed    ModuleVersionStatus = "reg_ingress_failed"
	ModuleVersionStatusOK                  ModuleVersionStatus = "ok"
)

var ErrInvalidModuleRepo = errors.New("invalid repository name for module")

type (
	Module struct {
		ID           string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		Name         string
		Provider     string
		Organization string // Module belongs to an organization
		Status       ModuleStatus
		Versions     []ModuleVersion  // versions sorted in descending order
		Connection   *repo.Connection // optional vcs repo connection
	}

	ModuleStatus string

	ModuleVersion struct {
		ID          string
		ModuleID    string
		Version     string
		CreatedAt   time.Time
		UpdatedAt   time.Time
		Status      ModuleVersionStatus
		StatusError string
		// TODO: download counters
	}

	ModuleVersionStatus string

	PublishModuleOptions struct {
		Repo          moduleRepo
		VCSProviderID string
	}
	PublishModuleVersionOptions struct {
		ModuleID string
		Version  string
		Ref      string
		Repo     moduleRepo
		Client   cloud.Client
	}
	CreateModuleOptions struct {
		Name         string
		Provider     string
		Organization string
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
	ListModulesOptions struct {
		Organization string `schema:"organization_name,required"` // filter by organization name
	}
	ModuleList struct {
		*otf.Pagination
		Items []*Module
	}
)

func NewModule(opts CreateModuleOptions) *Module {
	return &Module{
		ID:           otf.NewID("mod"),
		CreatedAt:    otf.CurrentTimestamp(),
		UpdatedAt:    otf.CurrentTimestamp(),
		Name:         opts.Name,
		Provider:     opts.Provider,
		Status:       ModuleStatusPending,
		Organization: opts.Organization,
	}
}

func NewModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		ID:        otf.NewID("modver"),
		CreatedAt: otf.CurrentTimestamp(),
		UpdatedAt: otf.CurrentTimestamp(),
		ModuleID:  opts.ModuleID,
		// TODO: check version is a semver, and decide whether to keep or drop
		// 'v' prefix
		Version: opts.Version,
		Status:  ModuleVersionStatusPending,
	}
}

func (m *Module) MarshalLog() any {
	return struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Provider     string `json:"provider"`
	}{
		ID:           m.ID,
		Organization: m.Organization,
		Name:         m.Name,
		Provider:     m.Provider,
	}
}

func (m *Module) AvailableVersions() (avail []ModuleVersion) {
	for _, modver := range m.Versions {
		if modver.Status == ModuleVersionStatusOK {
			avail = append(avail, modver)
		}
	}
	return
}

func (m *Module) Version(v string) *ModuleVersion {
	for _, modver := range m.Versions {
		if modver.Version == v {
			return &modver
		}
	}
	return nil
}

// Latest retrieves the latest version, which is the greatest version with an
// ok status. If there is such version, nil is returned.
func (m *Module) Latest() *ModuleVersion {
	for _, modver := range m.Versions {
		if modver.Status == ModuleVersionStatusOK {
			return &modver
		}
	}
	return nil
}
