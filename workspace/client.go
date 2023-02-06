package workspace

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (c *Client) CreateWorkspace(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error) {
	// Pre-emptively validate options
	if _, err := otf.NewWorkspace(opts); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/workspaces", *opts.Organization)
	req, err := c.NewRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// GetWorkspaceByName retrieves a workspace by organization and
// name.
func (c *Client) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*otf.Workspace, error) {
	path := fmt.Sprintf("organizations/%s/workspaces/%s", organization, workspace)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// GetWorkspace retrieves a workspace by its ID
func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *Client) ListWorkspaces(ctx context.Context, options otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*options.Organization))
	req, err := c.NewRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	wl := &jsonapi.WorkspaceList{}
	err = c.Do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceListJSONAPI(wl), nil
}

// UpdateWorkspace updates the settings of an existing workspace.
func (c *Client) UpdateWorkspace(ctx context.Context, workspaceID string, options otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	// Pre-emptively validate options
	if err := (&otf.Workspace{}).Update(options); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("workspaces/%s", workspaceID)
	req, err := c.NewRequest("PATCH", path, &jsonapi.WorkspaceUpdateOptions{
		ExecutionMode: (*string)(options.ExecutionMode),
	})
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *Client) LockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	path := fmt.Sprintf("workspaces/%s/actions/lock", workspaceID)
	req, err := c.NewRequest("POST", path, &opts)
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *Client) UnlockWorkspace(ctx context.Context, workspaceID string, _ otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	path := fmt.Sprintf("workspaces/%s/actions/unlock", workspaceID)
	req, err := c.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	w := &jsonapi.Workspace{}
	err = c.Do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}
