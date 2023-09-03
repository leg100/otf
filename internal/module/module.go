// Package module is reponsible for registry modules
package module

import (
	"errors"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/resource"
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
		Versions     []ModuleVersion         // versions sorted in descending order
		Connection   *connections.Connection // optional vcs repo connection
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

	PublishOptions struct {
		Repo          Repo
		VCSProviderID string
	}
	PublishVersionOptions struct {
		ModuleID string
		Version  string
		Ref      string
		Repo     Repo
		Client   cloud.Client
	}
	CreateOptions struct {
		Name         string
		Provider     string
		Organization string
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
		*resource.Pagination
		Items []*Module
	}
)

func newModule(opts CreateOptions) *Module {
	return &Module{
		ID:           internal.NewID("mod"),
		CreatedAt:    internal.CurrentTimestamp(),
		UpdatedAt:    internal.CurrentTimestamp(),
		Name:         opts.Name,
		Provider:     opts.Provider,
		Status:       ModuleStatusPending,
		Organization: opts.Organization,
	}
}

func newModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		ID:        internal.NewID("modver"),
		CreatedAt: internal.CurrentTimestamp(),
		UpdatedAt: internal.CurrentTimestamp(),
		ModuleID:  opts.ModuleID,
		// TODO: check version is a semver, and decide whether to keep or drop
		// 'v' prefix
		Version: opts.Version,
		Status:  ModuleVersionStatusPending,
	}
}

func (m *Module) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", m.ID),
		slog.String("organization", m.Organization),
		slog.String("name", m.Name),
		slog.String("provider", m.Provider),
		slog.String("status", string(m.Status)),
	}
	if m.Latest() != nil {
		attrs = append(attrs, slog.String("latest_version", m.Latest().Version))
	}
	return slog.GroupValue(attrs...)
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
// ok status. If there is no such version, nil is returned.
func (m *Module) Latest() *ModuleVersion {
	for _, modver := range m.Versions {
		if modver.Status == ModuleVersionStatusOK {
			return &modver
		}
	}
	return nil
}

func (v *ModuleVersion) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
		slog.String("module_id", v.ModuleID),
		slog.String("version", v.Version),
		slog.String("status", string(v.Status)),
	}
	return slog.GroupValue(attrs...)
}
