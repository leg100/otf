package module

import (
	"context"
)

type statusUpdater struct {
	*pgdb
}

// updateStatus updates a module version's status, and ensures its parent module's
// status reflects the version's status
func (u *statusUpdater) updateStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) (*otf.Module, *otf.ModuleVersion, error) {
	var mod *otf.Module
	var modver *otf.ModuleVersion

	err := u.tx(ctx, func(tx db) (err error) {
		// update version status
		modver, err = tx.UpdateModuleVersionStatus(ctx, opts)
		if err != nil {
			return err
		}

		// ensure module status reflects version status
		mod, err = tx.GetModuleByID(ctx, modver.ModuleID)
		if err != nil {
			return err
		}
		nextStatus := NextModuleStatus(mod.Status(), modver.Status())
		if nextStatus != mod.Status() {
			mod.UpdateStatus(nextStatus)
			err = tx.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
				ID:     mod.ID,
				Status: nextStatus,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return mod, modver, err
}
