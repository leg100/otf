package otf

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
		Connection   *Connection // optional vcs repo connection
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

	PublishModuleOptions struct {
		Repo          ModuleRepo
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
