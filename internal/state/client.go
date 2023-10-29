package state

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type Client struct {
	*otfapi.Client

	// Client does not implement all of service yet
	Service
}

func (c *Client) CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(*opts.WorkspaceID))
	req, err := c.NewRequest("POST", u, &types.StateVersionCreateVersionOptions{
		MD5:    internal.String(fmt.Sprintf("%x", md5.Sum(opts.State))),
		Serial: opts.Serial,
		State:  internal.String(base64.StdEncoding.EncodeToString(opts.State)),
	})
	if err != nil {
		return nil, err
	}

	var sv Version
	if err = c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}
	return &sv, nil
}

func (c *Client) ListStateVersions(ctx context.Context, workspaceID string, opts resource.PageOptions) (*resource.Page[*Version], error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(workspaceID))
	req, err := c.NewRequest("GET", u, &opts)
	if err != nil {
		return nil, err
	}
	var page resource.Page[*Version]
	if err := c.Do(ctx, req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error) {
	sv, err := c.GetCurrentStateVersion(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return c.DownloadState(ctx, sv.ID)
}

func (c *Client) GetCurrentStateVersion(ctx context.Context, workspaceID string) (*Version, error) {
	u := fmt.Sprintf("workspaces/%s/current-state-version", url.QueryEscape(workspaceID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var sv Version
	if err := c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}
	return &sv, nil
}

func (c *Client) DeleteStateVersion(ctx context.Context, svID string) error {
	u := fmt.Sprintf("state-versions/%s", url.QueryEscape(svID))
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	if err = c.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	u := fmt.Sprintf("state-versions/%s/download", url.QueryEscape(svID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Client) RollbackStateVersion(ctx context.Context, svID string) (*Version, error) {
	// The OTF JSON:API rollback endpoint matches the TFC endpoint for
	// compatibilty purposes, and takes both a workspace ID and a state version
	// ID, but OTF does nothing with the workspace ID and thus anything can be
	// specified.
	u := fmt.Sprintf("state-versions/%s/rollback", url.QueryEscape(svID))
	req, err := c.NewRequest("PATCH", u, nil)
	if err != nil {
		return nil, err
	}
	var sv Version
	if err = c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}
	return &sv, nil
}
