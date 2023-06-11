package workspace

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/apigen"
)

type Client struct {
	*otfhttp.Client
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

func (c *Client) ListWorkspaces(ctx context.Context, options ListOptions) (*WorkspaceList, error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*options.Organization))
	req, err := c.NewRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	wl := &types.WorkspaceList{}
	err = c.Do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return unmarshalListJSONAPI(wl), nil
}

// UpdateWorkspace updates the settings of an existing workspace.
func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID string, options UpdateOptions) (*Workspace, error) {
	var req apigen.UpdateWorkspace
	if options.ExecutionMode != nil {
		req.ExecutionMode = apigen.OptUpdateWorkspaceExecutionMode{
			Set:   true,
			Value: apigen.UpdateWorkspaceExecutionMode(*options.ExecutionMode),
		}
	}
	from, err := c.Client.UpdateWorkspace(ctx, req, apigen.UpdateWorkspaceParams{
		ID: workspaceID,
	})
	if err != nil {
		return nil, err
	}
	return c.toWorkspace(from), nil
}

func (c *Client) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*Workspace, error) {
	from, err := c.Client.LockWorkspace(ctx, apigen.ForceUnlockWorkspaceParams{
		ID: workspaceID,
	})
	if err != nil {
		return nil, err
	}
	return c.toWorkspace(from), nil
}

func (c *Client) UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*Workspace, error) {
	var (
		from *apigen.Workspace
		err  error
	)
	if force {
		from, err = c.Client.ForceUnlockWorkspace(ctx, apigen.ForceUnlockWorkspaceParams{
			ID: workspaceID,
		})
	} else {
		from, err = c.Client.UnlockWorkspace(ctx, apigen.UnlockWorkspaceParams{
			ID: workspaceID,
		})
	}
	if err != nil {
		return nil, err
	}
	return c.toWorkspace(from), nil
}

func (c *Client) toWorkspace(from *apigen.Workspace) *Workspace {
	return &Workspace{
		Name:      from.Name,
		ID:        from.ID,
		AutoApply: from.AutoApply,
	}
}
