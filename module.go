package otf

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/semver"
)

const (
	ModuleStatusPending       ModuleStatus = "pending"
	ModuleStatusNoVersionTags ModuleStatus = "no_version_tags"
	ModuleStatusSetupFailed   ModuleStatus = "setup_failed"
	ModuleStatusSetupComplete ModuleStatus = "setup_complete"
)

// ErrInvalidModuleName is returned when creating a module with a name that is
// invalid.
var ErrInvalidModuleName = errors.New("invalid module name")

type (
	ModuleStatus string

	Module struct {
		ID           string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		Name         string
		Provider     string
		Organization string // Module belongs to an organization
		Status       ModuleStatus
		Versions     map[string]*ModuleVersion
		Latest       *ModuleVersion
		Repo         *Connection // optional vcs repo connection
	}

	ModuleService interface {
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

	ModuleStore interface {
		CreateModule(context.Context, *Module) error
		UpdateModuleStatus(ctx context.Context, opts UpdateModuleStatusOptions) error
		ListModules(context.Context, ListModulesOptions) ([]*Module, error)
		GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error)
		GetModuleByID(ctx context.Context, id string) (*Module, error)
		GetModuleByWebhookID(ctx context.Context, id uuid.UUID) (*Module, error)
		DeleteModule(ctx context.Context, id string) error
	}
	PublishModuleOptions struct {
		Identifier    string
		VCSProviderID string
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
		*Pagination
		Items []*Module
	}
)

func NewModule(opts CreateModuleOptions) *Module {
	return &Module{
		ID:           NewID("mod"),
		CreatedAt:    CurrentTimestamp(),
		UpdatedAt:    CurrentTimestamp(),
		Name:         opts.Name,
		Provider:     opts.Provider,
		Status:       ModuleStatusPending,
		Organization: opts.Organization,
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

func (m *Module) AvailableVersions() (avail []*ModuleVersion) {
	for _, v := range m.Versions {
		if v.Status == ModuleVersionStatusOK {
			avail = append(avail, v)
		}
	}
	return
}

func (m *Module) AddVersion(modver *ModuleVersion) error {
	if _, ok := m.Versions[modver.Version]; ok {
		return errors.New("version already taken")
	}
	m.Versions[modver.Version] = modver

	return nil
}

// RemoveVersion removes a module version by its ID.
func (m *Module) RemoveVersion(versionID string) error {
	if _, ok := m.Versions[versionID]; !ok {
		return errors.New("version not found")
	}
	delete(m.Versions, versionID)

	return nil
}

// setLatest sets the latest version, which is the greatest version with an
// ok status. If a new latest version is set then true is returned.
func (m *Module) SetLatest() bool {
	var currentID *string // ID of current latest version
	if m.Latest != nil {
		currentID = &m.Latest.ID
	}

	versions := make([]string, len(m.Versions))
	for v := range m.Versions {
		versions = append(versions, v)
	}
	semver.Sort(versions)
	for i := len(versions) - 1; i >= 0; i-- {
		if m.Versions[versions[i]].Status == ModuleVersionStatusOK {
			m.Latest = m.Versions[versions[i]]
			return currentID != nil && *currentID != m.Latest.ID
		}
	}
	// no ok version found
	m.Latest = nil
	return currentID != nil
}
