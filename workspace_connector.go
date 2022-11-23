package otf

import (
	"context"
	"errors"
)

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// certain VCS events on the repo, and triggering runs accordingly.
type WorkspaceConnector struct {
	Host string // otfd host, to which webhook events are sent

	Application
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	HTTPURL    string `schema:"http_url,required"`   // complete HTTP/S URL for repo
	ProviderID string `schema:"vcs_provider_id,required"`
	Branch     string `schema:"branch,required"`
	// otfd's externally-facing host[:port], the destination for VCS events
	Host string
}

func (wc *WorkspaceConnector) Connect(ctx context.Context, spec WorkspaceSpec, opts ConnectWorkspaceOptions) (*Workspace, error) {
	webhook, err := NewWebhook(opts.Identifier, opts.HTTPURL)
	if err != nil {
		return nil, err
	}

	// Perform three operations within a transaction:
	// 1. get or create webhook in db
	// 2. create workspace repo in db
	// 3. create webhook on vcs provider
	var ws *Workspace
	err = wc.Tx(ctx, func(app Application) (err error) {
		webhook, err = app.DB().GetOrCreateWebhook(ctx, webhook)
		if err != nil {
			return err
		}

		ws, err = app.DB().CreateWorkspaceRepo(ctx, spec, WorkspaceRepo{
			Branch:     opts.Branch,
			ProviderID: opts.ProviderID,
			Webhook:    webhook,
		})
		if err != nil {
			return err
		}

		err = app.CreateWebhook(ctx, opts.ProviderID, CreateCloudWebhookOptions{
			Identifier: opts.Identifier,
			Secret:     webhook.Secret,
			Host:       opts.Host,
			WebhookID:  webhook.WebhookID,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ws, nil
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
		if errors.Is(err, ErrResourceReferenceViolation) {
			// webhook is still in use by another workspace
			return nil
		} else if err != nil {
			return err
		}

		err = app.DeleteWebhook(ctx, repo.ProviderID, repo.Webhook)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ws, nil
}
