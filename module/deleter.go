package module

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

// deleter deletes registry modules
type deleter struct {
	otf.HookService        // for unhooking from webhook
	otf.ModuleService      // for retrieving and deleting module
	otf.VCSProviderService // for retrieving cloud client
}

func NewDeleter(app otf.Application) *deleter {
	return &deleter{
		HookService:        app,
		ModuleService:      app,
		VCSProviderService: app,
	}
}

// Delete deletes the module and unhooks it from a webhook
func (c *deleter) Delete(ctx context.Context, moduleID string) error {
	ws, err := c.GetModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	if ws.Repo() == nil {
		// there is no webhook to unhook from, so just delete the module
		_, err = c.DeleteModule(ctx, moduleID)
		return err
	}

	client, err := c.GetVCSClient(ctx, ws.Repo().ProviderID)
	if err != nil {
		return err
	}

	unhookCallback := func(ctx context.Context, tx otf.Database) error {
		return sql.DeleteModule(ctx, tx, moduleID)
	}
	return c.Unhook(ctx, otf.UnhookOptions{
		HookID:         ws.Repo().WebhookID,
		Client:         client,
		UnhookCallback: unhookCallback,
	})
}
