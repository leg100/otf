package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/pkg/errors"
)

type Connector struct {
	otf.HookService        // for registering and unregistering connections to webhooks
	otf.VCSProviderService // for retrieving cloud client
}

func (c *Connector) Connect(ctx context.Context, workspaceID string, opts otf.ConnectWorkspaceOptions) error {
	client, err := c.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return err
	}

	repo, err := client.GetRepository(ctx, opts.Identifier)
	if err != nil {
		return errors.Wrap(err, "retrieving repository info")
	}

	hookCallback := func(ctx context.Context, tx otf.Database, hookID uuid.UUID) error {
		_, err := newPGDB(tx).CreateWorkspaceRepo(ctx, workspaceID, WorkspaceRepo{
			Branch:     repo.Branch,
			ProviderID: opts.ProviderID,
			WebhookID:  hookID,
		})
		return err
	}
	err = c.Hook(ctx, otf.HookOptions{
		Identifier:   opts.Identifier,
		Cloud:        opts.Cloud,
		HookCallback: hookCallback,
		Client:       client,
	})
	if err != nil {
		return errors.Wrap(err, "registering webhook connection")
	}
	return nil
}

// Disconnect a repo from a workspace. The repo's webhook is deleted if no other
// workspace is connected to the repo.
func (c *Connector) Disconnect(ctx context.Context, ws *Workspace) (*Workspace, error) {
	repo := ws.Repo()
	client, err := c.GetVCSClient(ctx, repo.ProviderID)
	if err != nil {
		return nil, err
	}

	// Perform multiple operations within a transaction:
	// 1. delete workspace repo from db
	// 2. delete webhook from db
	unhookCallback := func(ctx context.Context, tx otf.Database) error {
		ws, err = newPGDB(tx).DeleteWorkspaceRepo(ctx, ws.id)
		return err
	}
	err = c.Unhook(ctx, otf.UnhookOptions{
		HookID:         repo.WebhookID,
		Client:         client,
		UnhookCallback: unhookCallback,
	})
	return ws, err
}
