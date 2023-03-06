package otf

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	repo         *Connection // Module optionally connected to vcs repo
	status       ModuleStatus
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
	Delete(ctx context.Context, module *Module) error
}

type (
	PublishModuleOptions struct {
		Identifier   string
		ProviderID   string
		Organization string
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
