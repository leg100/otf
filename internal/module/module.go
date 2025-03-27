// Package module is reponsible for registry modules
package module

import (
	"errors"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
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
		ID           resource.TfeID `db:"module_id"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
		Name         string
		Provider     string
		Status       ModuleStatus
		Organization resource.OrganizationName `db:"organization_name"` // Module belongs to an organization
		Versions     []ModuleVersion           `db:"module_versions"`   // versions sorted in descending order
		Connection   *connections.Connection   // optional vcs repo connection
	}

	ModuleStatus string

	// ModuleVersion is a version of a module.
	//
	// NOTE: field order must match postgres table column order.
	ModuleVersion struct {
		ID          resource.TfeID `db:"module_version_id"`
		Version     string
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
		Status      ModuleVersionStatus
		StatusError *string        `db:"status_error"`
		ModuleID    resource.TfeID `db:"module_id"`

		// TODO: download counters
	}

	ModuleVersionStatus string

	PublishOptions struct {
		Repo          Repo
		VCSProviderID resource.TfeID
	}
	PublishVersionOptions struct {
		ModuleID resource.TfeID
		Version  string
		Ref      string
		Repo     Repo
		Client   vcs.Client
	}
	CreateOptions struct {
		Name         string
		Provider     string
		Organization resource.OrganizationName
	}
	CreateModuleVersionOptions struct {
		ModuleID resource.TfeID
		Version  string
	}
	UpdateModuleVersionStatusOptions struct {
		ID     resource.TfeID
		Status ModuleVersionStatus
		Error  string
	}
	GetModuleOptions struct {
		Name         string                    `schema:"name,required"`
		Provider     string                    `schema:"provider,required"`
		Organization resource.OrganizationName `schema:"organization,required"`
	}
	ListModulesOptions struct {
		Organization resource.OrganizationName `schema:"organization_name,required"` // filter by organization name
	}
	ModuleList struct {
		*resource.Pagination
		Items []*Module
	}
)

func newModule(opts CreateOptions) *Module {
	return &Module{
		ID:           resource.NewTfeID(resource.ModuleKind),
		CreatedAt:    internal.CurrentTimestamp(nil),
		UpdatedAt:    internal.CurrentTimestamp(nil),
		Name:         opts.Name,
		Provider:     opts.Provider,
		Status:       ModuleStatusPending,
		Organization: opts.Organization,
	}
}

func newModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		ID:        resource.NewTfeID(resource.ModuleVersionKind),
		CreatedAt: internal.CurrentTimestamp(nil),
		UpdatedAt: internal.CurrentTimestamp(nil),
		ModuleID:  opts.ModuleID,
		// TODO: check version is a semver, and decide whether to keep or drop
		// 'v' prefix
		Version: opts.Version,
		Status:  ModuleVersionStatusPending,
	}
}

func (m *Module) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", m.ID.String()),
		slog.Any("organization", m.Organization),
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
		slog.String("id", v.ID.String()),
		slog.String("module_id", v.ModuleID.String()),
		slog.String("version", v.Version),
		slog.String("status", string(v.Status)),
	}
	return slog.GroupValue(attrs...)
}
