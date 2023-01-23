package otf

import (
	"context"

	"github.com/pkg/errors"
)

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type WorkspaceConnector struct {
	Application
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud host of the repo
}

func (wc *WorkspaceConnector) Connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) (*Workspace, error) {
	client, err := wc.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return nil, err
	}

	repo, err := client.GetRepository(ctx, opts.Identifier)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving repository info")
	}

	// Inside transaction:
	// 1. create webhook
	// 2. create workspace repo in store
	var ws *Workspace
	err = wc.Tx(ctx, func(app Application) (err error) {
		webhook, err := app.CreateWebhook(ctx, CreateWebhookOptions(opts))
		if err != nil {
			return errors.Wrap(err, "creating webhook")
		}

		ws, err = app.DB().CreateWorkspaceRepo(ctx, workspaceID, WorkspaceRepo{
			Branch:     repo.Branch,
			ProviderID: opts.ProviderID,
			WebhookID:  webhook.ID(),
		})
		return errors.Wrap(err, "creating workspace repo")
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}
	return ws, nil
}

// Disconnect a repo from a workspace. The repo's webhook is deleted if no other
// workspace is connected to the repo.
func (wc *WorkspaceConnector) Disconnect(ctx context.Context, workspaceID string) (*Workspace, error) {
	ws, err := wc.DB().GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	repo := ws.Repo()
	// Perform multiple operations within a transaction:
	// 1. delete workspace repo from db
	// 2. delete webhook from db
	err = wc.Tx(ctx, func(app Application) (err error) {
		ws, err = app.DB().DeleteWorkspaceRepo(ctx, workspaceID)
		if err != nil {
			return err
		}

		return app.DeleteWebhook(ctx, repo.ProviderID, repo.WebhookID)
	})
	return ws, err
}
