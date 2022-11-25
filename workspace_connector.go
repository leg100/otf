package otf

import (
	"context"
	"errors"
	"reflect"
)

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// certain VCS events on the repo, and triggering runs accordingly.
type WorkspaceConnector struct {
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
	// create webhook obj and attempt to persist it to the db
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}
	webhook := &Webhook{
		Secret: secret,
	}

	// func for ensuring vcs provider's webhook config matches config in otf.
	sync := func(hook *Webhook) (string, error) {
		if hook.VCSID == nil {
			return wc.CreateWebhook(ctx, opts.ProviderID, CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     hook.Secret,
				Events:     WebhookEvents,
				URL:        opts.Host,
			})
		}

		vcsHook, err := wc.GetWebhook(ctx, opts.ProviderID, GetWebhookOptions{
			Identifier: opts.Identifier,
			ID:         *hook.VCSID,
		})
		if errors.Is(err, ErrResourceNotFound) {
			return wc.CreateWebhook(ctx, opts.ProviderID, CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     hook.Secret,
				Events:     WebhookEvents,
				URL:        opts.Host,
			})
		} else if err != nil {
			return "", err
		}

		if !reflect.DeepEqual(vcsHook.Events, WebhookEvents) ||
			vcsHook.Endpoint != opts.Host {

			err = wc.UpdateWebhook(ctx, opts.ProviderID, UpdateWebhookOptions{
				ID: *hook.VCSID,
				CreateWebhookOptions: CreateWebhookOptions{
					Identifier: opts.Identifier,
					Secret:     hook.Secret,
					URL:        opts.Host,
				},
			})
			return *hook.VCSID, err
		}
		return *hook.VCSID, nil
	}

	// Inside transaction:
	// 1. synchronise webhook config
	// 2. create workspace repo in store
	var ws *Workspace
	err = wc.Tx(ctx, func(app Application) (err error) {
		webhook, err = app.DB().SyncWebhook(ctx, webhook, sync)
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
		if errors.Is(err, ErrResourceReferenceViolation) {
			// webhook is still in use by another workspace
			return nil
		} else if err != nil {
			return err
		}

		err = app.DeleteWebhook(ctx, repo.ProviderID, DeleteWebhookOptions{
			Identifier: repo.Identifier,
			ID:         *repo.Webhook.VCSID,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return ws, err
}
