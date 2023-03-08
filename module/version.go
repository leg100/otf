package module

import (
	"context"

	"github.com/leg100/otf"
)

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

		mod, err = tx.DB().GetModuleByID(ctx, modver.ModuleID)
		if err != nil {
			return err
		}
		newModStatus := NextModuleStatus(mod.Status(), modver.Status())
		if newModStatus != mod.Status() {
			mod, err = tx.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
				ID:     mod.ID,
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
