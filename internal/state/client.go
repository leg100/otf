package state

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/url"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfhttp.Client
}

func (c *Client) Create(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(opts.WorkspaceID.String()))
	req, err := c.NewRequest("POST", u, &TFEStateVersionCreateVersionOptions{
		MD5:    new(fmt.Sprintf("%x", md5.Sum(opts.State))),
		Serial: opts.Serial,
		State:  new(base64.StdEncoding.EncodeToString(opts.State)),
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

func (c *Client) List(ctx context.Context, workspaceID resource.TfeID, opts resource.PageOptions) (*resource.Page[*Version], error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(workspaceID.String()))
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

func (c *Client) DownloadCurrent(ctx context.Context, workspaceID resource.TfeID) ([]byte, error) {
	sv, err := c.GetCurrent(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return c.Download(ctx, sv.ID)
}

func (c *Client) GetCurrent(ctx context.Context, workspaceID resource.TfeID) (*Version, error) {
	u := fmt.Sprintf("workspaces/%s/current-state-version", url.QueryEscape(workspaceID.String()))
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

func (c *Client) Delete(ctx context.Context, svID resource.TfeID) error {
	u := fmt.Sprintf("state-versions/%s", url.QueryEscape(svID.String()))
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	if err = c.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) Download(ctx context.Context, svID resource.TfeID) ([]byte, error) {
	u := fmt.Sprintf("state-versions/%s/download", url.QueryEscape(svID.String()))
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

func (c *Client) Rollback(ctx context.Context, svID resource.TfeID) (*Version, error) {
	// The OTF JSON:API rollback endpoint matches the TFC endpoint for
	// compatibilty purposes, and takes both a workspace ID and a state version
	// ID, but OTF does nothing with the workspace ID and thus anything can be
	// specified.
	u := fmt.Sprintf("state-versions/%s/rollback", url.QueryEscape(svID.String()))
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
