package module

import (
	"context"

	"github.com/leg100/otf"
)

// deleter deletes registry modules
type deleter struct {
	otf.RepoService // for disconnecting from repo
	otf.DB          // for deleting module from db
}

func NewDeleter(app otf.Application) *deleter {
	return &deleter{
		RepoService: app,
		DB:          app.DB(),
	}
}

// Delete deletes the module and disconnects it from a VCS repo.
func (c *deleter) Delete(ctx context.Context, mod *otf.Module) error {
	return c.Tx(ctx, func(tx otf.DB) error {
		if err := tx.DeleteModule(ctx, mod.ID()); err != nil {
			return err
		}

		if mod.Repo() == nil {
			return nil // not connected; skip disconnection
		}

		return c.RepoService.Disconnect(ctx, otf.DisconnectOptions{
			ConnectionType: otf.ModuleConnection,
			ResourceID:     mod.ID(),
			Tx:             tx,
		})
	})
}
