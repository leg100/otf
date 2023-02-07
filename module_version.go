package otf

import (
	"context"
	"time"
)

type ModuleVersionStatus string

// List of available registry module version statuses
const (
	ModuleVersionStatusPending             ModuleVersionStatus = "pending"
	ModuleVersionStatusCloning             ModuleVersionStatus = "cloning"
	ModuleVersionStatusCloneFailed         ModuleVersionStatus = "clone_failed"
	ModuleVersionStatusRegIngressReqFailed ModuleVersionStatus = "reg_ingress_req_failed"
	ModuleVersionStatusRegIngressing       ModuleVersionStatus = "reg_ingressing"
	ModuleVersionStatusRegIngressFailed    ModuleVersionStatus = "reg_ingress_failed"
	ModuleVersionStatusOk                  ModuleVersionStatus = "ok"
)

type ModuleVersion interface {
	ID() string
	ModuleID() string
	Version() string
	Status() ModuleVersionStatus
	StatusError() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type ModuleVersionService interface {
	CreateModuleVersion(context.Context, CreateModuleVersionOptions) (*ModuleVersion, error)
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) (*ModuleVersion, error)
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
}

type ModuleVersionStore interface {
	CreateModuleVersion(context.Context, *ModuleVersion) error
	UpdateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) (*ModuleVersion, error)
	UploadModuleVersion(ctx context.Context, opts UploadModuleVersionOptions) error
	DownloadModuleVersion(ctx context.Context, opts DownloadModuleOptions) ([]byte, error)
}
