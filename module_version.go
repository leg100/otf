package otf

import "time"

type ModuleVersion struct {
	id                         string
	moduleID                   string
	version                    string
	createdAt                  time.Time
	updatedAt                  time.Time
	inputs, outputs, resources int
	readme                     string
	// TODO: download counter
}

func NewModuleVersion(opts CreateModuleVersionOptions) *ModuleVersion {
	return &ModuleVersion{
		id:        NewID("mod-ver"),
		createdAt: CurrentTimestamp(),
		updatedAt: CurrentTimestamp(),
		moduleID:  opts.ModuleID,
		version:   opts.Version,
	}
}

func (v *ModuleVersion) ID() string           { return v.id }
func (v *ModuleVersion) ModuleID() string     { return v.moduleID }
func (v *ModuleVersion) Version() string      { return v.version }
func (v *ModuleVersion) CreatedAt() time.Time { return v.createdAt }
func (v *ModuleVersion) UpdatedAt() time.Time { return v.updatedAt }

func (v *ModuleVersion) MarshalLog() any {
	return struct {
		ID, ModuleID, Version string
	}{
		ID:       v.id,
		ModuleID: v.moduleID,
		Version:  v.version,
	}
}
