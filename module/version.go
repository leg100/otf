package module

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
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

type ModuleVersion struct {
	id          string
	moduleID    string
	version     string
	createdAt   time.Time
	updatedAt   time.Time
	status      ModuleVersionStatus
	statusError string
	// TODO: download counter
}

func NewModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		id:        otf.NewID("modver"),
		createdAt: otf.CurrentTimestamp(),
		updatedAt: otf.CurrentTimestamp(),
		moduleID:  opts.ModuleID,
		version:   opts.Version,
		status:    ModuleVersionStatusPending,
	}
}

func (v *ModuleVersion) ID() string                  { return v.id }
func (v *ModuleVersion) ModuleID() string            { return v.moduleID }
func (v *ModuleVersion) Version() string             { return v.version }
func (v *ModuleVersion) Status() ModuleVersionStatus { return v.status }
func (v *ModuleVersion) StatusError() string         { return v.statusError }
func (v *ModuleVersion) CreatedAt() time.Time        { return v.createdAt }
func (v *ModuleVersion) UpdatedAt() time.Time        { return v.updatedAt }

func (v *ModuleVersion) MarshalLog() any {
	return struct {
		ID, ModuleID, Version string
	}{
		ID:       v.id,
		ModuleID: v.moduleID,
		Version:  v.version,
	}
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

// UpdateModuleVersionStatus updates a module version's status, and ensures its parent module's
// status is updated accordingly.
func UpdateModuleVersionStatus(ctx context.Context, app otf.Application, opts UpdateModuleVersionStatusOptions) (*Module, *ModuleVersion, error) {
	var mod *Module
	var modver *ModuleVersion

	err := app.Tx(ctx, func(tx otf.Application) (err error) {
		modver, err = tx.DB().UpdateModuleVersionStatus(ctx, opts)
		if err != nil {
			return err
		}

		mod, err = tx.DB().GetModuleByID(ctx, modver.ModuleID())
		if err != nil {
			return err
		}
		newModStatus := NextModuleStatus(mod.Status(), modver.Status())
		if newModStatus != mod.Status() {
			mod, err = tx.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
				ID:     mod.ID(),
				Status: newModStatus,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return mod, modver, err
}

// ModuleVersionRow is a module version database row
type ModuleVersionRow struct {
	ModuleVersionID pgtype.Text        `json:"module_version_id"`
	Version         pgtype.Text        `json:"version"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
	Status          pgtype.Text        `json:"status"`
	StatusError     pgtype.Text        `json:"status_error"`
	ModuleID        pgtype.Text        `json:"module_id"`
}

// UnmarshalModuleVersionRow unmarshals a database row into a module version
func UnmarshalModuleVersionRow(row ModuleVersionRow) *ModuleVersion {
	return &ModuleVersion{
		id:          row.ModuleVersionID.String,
		version:     row.Version.String,
		createdAt:   row.CreatedAt.Time.UTC(),
		updatedAt:   row.UpdatedAt.Time.UTC(),
		moduleID:    row.ModuleID.String,
		status:      ModuleVersionStatus(row.Status.String),
		statusError: row.StatusError.String,
	}
}
