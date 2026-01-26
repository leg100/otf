package workspace

import (
	"context"
	"fmt"
	"net/url"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfhttp.Client
}

func (c *Client) GetByName(ctx context.Context, organization organization.Name, workspace string) (*Workspace, error) {
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

func (c *Client) Get(ctx context.Context, workspaceID resource.TfeID) (*Workspace, error) {
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

func (c *Client) List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(opts.Organization.String()))
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

func (c *Client) Update(ctx context.Context, workspaceID resource.TfeID, opts UpdateOptions) (*Workspace, error) {
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

func (c *Client) Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*Workspace, error) {
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

func (c *Client) Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*Workspace, error) {
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
