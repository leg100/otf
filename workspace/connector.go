package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/vcsprovider"
	"github.com/pkg/errors"
)

// connector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type Connector struct {
	otf.HookService      // for registering and unregistering connections to webhooks
	*vcsprovider.Service // for retrieving cloud client
}

type connectOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud host of the repo
}

func (c *Connector) connect(ctx context.Context, workspaceID string, opts connectOptions) error {
	client, err := c.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return err
	}

	repo, err := client.GetRepository(ctx, opts.Identifier)
	if err != nil {
		return errors.Wrap(err, "retrieving repository info")
	}

	hookCallback := func(ctx context.Context, tx otf.DB, hookID uuid.UUID) error {
		_, err := newdb(tx).CreateWorkspaceRepo(ctx, workspaceID, WorkspaceRepo{
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
func (c *Connector) disconnect(ctx context.Context, ws *Workspace) (*Workspace, error) {
	repo := ws.Repo()
	client, err := c.GetVCSClient(ctx, repo.ProviderID)
	if err != nil {
		return nil, err
	}

	// Perform multiple operations within a transaction:
	// 1. delete workspace repo from db
	// 2. delete webhook from db
	unhookCallback := func(ctx context.Context, tx otf.DB) error {
		ws, err = newdb(tx).DeleteWorkspaceRepo(ctx, ws.id)
		return err
	}
	err = c.Unhook(ctx, otf.UnhookOptions{
		HookID:         repo.WebhookID,
		Client:         client,
		UnhookCallback: unhookCallback,
	})
	return ws, err
}
