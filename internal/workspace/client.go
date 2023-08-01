package workspace

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	internal.JSONAPIClient
}

// GetWorkspaceByName retrieves a workspace by organization and
// name.
func (c *Client) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	path := fmt.Sprintf("organizations/%s/workspaces/%s", organization, workspace)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return unmarshalJSONAPI(w), nil
}

// GetWorkspace retrieves a workspace by its ID
func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return unmarshalJSONAPI(w), nil
}

func (c *Client) ListWorkspaces(ctx context.Context, options ListOptions) (*resource.Page[*Workspace], error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*options.Organization))
	req, err := c.NewRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	wl := &types.WorkspaceList{}
	if err = c.Do(ctx, req, wl); err != nil {
		return nil, err
	}

	return unmarshalListJSONAPI(wl), nil
}

// UpdateWorkspace updates the settings of an existing workspace.
func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID string, options UpdateOptions) (*Workspace, error) {
	// Pre-emptively validate options
	if _, err := (&Workspace{}).Update(options); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("PATCH", path, &types.WorkspaceUpdateOptions{
		ExecutionMode: (*string)(options.ExecutionMode),
	})
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return unmarshalJSONAPI(w), nil
}

func (c *Client) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*Workspace, error) {
	path := fmt.Sprintf("workspaces/%s/actions/lock", workspaceID)
	req, err := c.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return unmarshalJSONAPI(w), nil
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

	w := &types.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return unmarshalJSONAPI(w), nil
}
