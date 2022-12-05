package otf

import (
	"context"
	"errors"
)

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type WorkspaceConnector struct {
	Application
	*WebhookCreator
	*WebhookUpdater
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	HTTPURL    string `schema:"http_url,required"`   // complete HTTP/S URL for repo
	ProviderID string `schema:"vcs_provider_id,required"`
	Branch     string `schema:"branch,required"`
	Cloud      string // cloud host of the repo
	OTFHost    string // externally-facing host[:port], the destination for VCS events
}

func (wc *WorkspaceConnector) Connect(ctx context.Context, spec WorkspaceSpec, opts ConnectWorkspaceOptions) (*Workspace, error) {
	// Inside transaction:
	// 1. synchronise webhook config
	// 2. create workspace repo in store
	var ws *Workspace
	err := wc.Tx(ctx, func(app Application) (err error) {
		webhook, err := app.DB().SyncWebhook(ctx, SyncWebhookOptions{
			Identifier:        opts.Identifier,
			HTTPURL:           opts.HTTPURL,
			ProviderID:        opts.ProviderID,
			OTFHost:           opts.OTFHost,
			Cloud:             opts.Cloud,
			CreateWebhookFunc: wc.Create,
			UpdateWebhookFunc: wc.Update,
		})
		if err != nil {
			return err
		}

		ws, err = app.DB().CreateWorkspaceRepo(ctx, spec, WorkspaceRepo{
			Branch:     opts.Branch,
			ProviderID: opts.ProviderID,
			Webhook:    webhook,
		})
		return err
	})
	return ws, err
}

// Disconnect a repo from a workspace. The repo's webhook is deleted if no other
// workspace is connected to the repo.
func (wc *WorkspaceConnector) Disconnect(ctx context.Context, spec WorkspaceSpec) (*Workspace, error) {
	// Perform three operations within a transaction:
	// 1. delete workspace repo from db
	// 2. delete webhook from db
	// 3. delete webhook from vcs provider
	var ws *Workspace
	err := wc.Tx(ctx, func(app Application) (err error) {
		ws, err = app.DB().GetWorkspace(ctx, spec)
		if err != nil {
			return err
		}
		repo := ws.Repo()

		ws, err = app.DB().DeleteWorkspaceRepo(ctx, spec)
		if err != nil {
			return err
		}

		err = app.DB().DeleteWebhook(ctx, repo.WebhookID)
		if errors.Is(err, ErrForeignKeyViolation) {
			// webhook is still in use by another workspace
			return nil
		} else if err != nil {
			return err
		}

		err = app.DeleteWebhook(ctx, repo.ProviderID, DeleteWebhookOptions{
			Identifier: repo.Identifier,
			ID:         repo.VCSID,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return ws, err
}
