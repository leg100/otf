package workspace

import (
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	internal.JSONAPIClient

	WorkspaceService
}

func (c *Client) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	path := fmt.Sprintf("organizations/%s/workspaces/%s", organization, workspace)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*opts.Organization))
	req, err := c.NewRequest("GET", u, &opts)
	if err != nil {
		return nil, err
	}
	var page resource.Page[*Workspace]
	if err = c.Do(ctx, req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateOptions) (*Workspace, error) {
	// Pre-emptively validate options
	if _, err := (&Workspace{}).Update(opts); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("PATCH", path, &opts)
	if err != nil {
		return nil, err
	}

	var ws Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (c *Client) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*Workspace, error) {
	path := fmt.Sprintf("workspaces/%s/actions/lock", workspaceID)
	req, err := c.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	var ws Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

func (c *Client) UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*Workspace, error) {
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

	var ws Workspace
	if err := c.Do(ctx, req, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}
