package otf

import (
	"time"
)

// List of available registry module version statuses
const (
	ModuleVersionStatusPending             ModuleVersionStatus = "pending"
	ModuleVersionStatusCloning             ModuleVersionStatus = "cloning"
	ModuleVersionStatusCloneFailed         ModuleVersionStatus = "clone_failed"
	ModuleVersionStatusRegIngressReqFailed ModuleVersionStatus = "reg_ingress_req_failed"
	ModuleVersionStatusRegIngressing       ModuleVersionStatus = "reg_ingressing"
	ModuleVersionStatusRegIngressFailed    ModuleVersionStatus = "reg_ingress_failed"
	ModuleVersionStatusOK                  ModuleVersionStatus = "ok"
)

type (
	ModuleVersionStatus string

	ModuleVersion struct {
		ID          string
		ModuleID    string
		Version     string
		CreatedAt   time.Time
		UpdatedAt   time.Time
		Status      ModuleVersionStatus
		StatusError string
		// TODO: download counter
	}
)

func NewModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		ID:        NewID("modver"),
		CreatedAt: CurrentTimestamp(),
		UpdatedAt: CurrentTimestamp(),
		ModuleID:  opts.ModuleID,
		Version:   opts.Version,
		Status:    ModuleVersionStatusPending,
	}
}

func (v *ModuleVersion) MarshalLog() any {
	return struct {
		ID, ModuleID, Version string
	}{
		ID:       v.ID,
		ModuleID: v.ModuleID,
		Version:  v.Version,
	}
}
