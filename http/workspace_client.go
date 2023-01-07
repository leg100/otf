package http

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (c *client) CreateWorkspace(ctx context.Context, options otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(options.OrganizationName))
	req, err := c.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = c.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// GetWorkspace retrieves a workspace either by its ID, or by organization and
// workspace name.
func (c *client) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	path, err := getWorkspacePath(spec)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = c.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *client) ListWorkspaces(ctx context.Context, options otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*options.OrganizationName))
	req, err := c.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	wl := &dto.WorkspaceList{}
	err = c.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceListJSONAPI(wl), nil
}

// UpdateWorkspace updates the settings of an existing workspace.
func (c *client) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, options otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	path, err := getWorkspacePath(spec)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest("PATCH", path, &dto.WorkspaceUpdateOptions{
		ExecutionMode: (*string)(options.ExecutionMode),
	})
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = c.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *client) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	var path string
	if spec.ID != nil {
		path = fmt.Sprintf("workspaces/%s/actions/lock", url.QueryEscape(*spec.ID))
	} else if spec.OrganizationName != nil && spec.Name != nil {
		path = fmt.Sprintf("organizations/%s/workspaces/%s/actions/lock", url.QueryEscape(*spec.OrganizationName), url.QueryEscape(*spec.Name))
	} else {
		return nil, otf.ErrInvalidWorkspaceSpec
	}
	req, err := c.newRequest("POST", path, &opts)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = c.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *client) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	var path string
	if spec.ID != nil {
		path = fmt.Sprintf("workspaces/%s/actions/unlock", url.QueryEscape(*spec.ID))
	} else if spec.OrganizationName != nil && spec.Name != nil {
		path = fmt.Sprintf("organizations/%s/workspaces/%s/actions/unlock", url.QueryEscape(*spec.OrganizationName), url.QueryEscape(*spec.Name))
	} else {
		return nil, otf.ErrInvalidWorkspaceSpec
	}
	req, err := c.newRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = c.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

func (c *client) GetWorkspaceQueue(workspaceID string) ([]*otf.Run, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *client) UpdateWorkspaceQueue(run *otf.Run) error {
	return fmt.Errorf("unimplemented")
}

func (c *client) SetLatestRun(ctx context.Context, workspaceID, runID string) error {
	return fmt.Errorf("unimplemented")
}

// getWorkspacePath generates a URL path for a workspace according to whether
// the spec specifies an ID, or an organization and workspace name.
func getWorkspacePath(spec otf.WorkspaceSpec) (string, error) {
	if spec.ID != nil {
		return fmt.Sprintf("workspaces/%s", url.QueryEscape(*spec.ID)), nil
	}

	if spec.Name != nil && spec.OrganizationName != nil {
		return fmt.Sprintf(
			"organizations/%s/workspaces/%s",
			url.QueryEscape(*spec.OrganizationName),
			url.QueryEscape(*spec.Name),
		), nil
	}

	return "", fmt.Errorf("invalid workspace spec")
}
