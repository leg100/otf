package api

import (
	"context"
	"fmt"
	"net/url"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

// Alias client to permit embedding it with other clients in a struct
// without a name clash.
type WorkspaceClient = Client

type Client struct {
	*otfhttp.Client
}

func (c *Client) GetWorkspaceByName(ctx context.Context, organization organization.Name, name string) (*workspace.Workspace, error) {
	path := fmt.Sprintf("organizations/%s/workspaces/%s", organization, name)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var ws workspace.Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) GetWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error) {
	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var ws workspace.Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(opts.Organization.String()))
	req, err := c.NewRequest("GET", u, &opts)
	if err != nil {
		return nil, err
	}
	var page resource.Page[*workspace.Workspace]
	if err = c.Do(ctx, req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID resource.TfeID, opts workspace.UpdateOptions) (*workspace.Workspace, error) {
	// Pre-emptively validate options
	if _, err := (&workspace.Workspace{}).Update(opts); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("PATCH", path, &opts)
	if err != nil {
		return nil, err
	}

	var ws workspace.Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (c *Client) Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*workspace.Workspace, error) {
	path := fmt.Sprintf("workspaces/%s/actions/lock", workspaceID)
	req, err := c.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	var ws workspace.Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (c *Client) Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error) {
	var u string
	if force {
		u = fmt.Sprintf("workspaces/%s/actions/unlock", workspaceID)
	} else {
		u = fmt.Sprintf("workspaces/%s/actions/force-unlock", workspaceID)
	}
	req, err := c.NewRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}

	var ws workspace.Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}
