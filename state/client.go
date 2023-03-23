package state

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// Client uses json-api according to the documented terraform cloud state
// version API [1] that OTF implements (we could use something different,
// something simpler but since the terraform CLI talks to OTF via json-api we
// might as well use this too...).
//
// [1] https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
type Client struct {
	otf.JSONAPIClient
}

func (c *Client) CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	var state file
	if err := json.Unmarshal(opts.State, &state); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(*opts.WorkspaceID))
	req, err := c.NewRequest("POST", u, &jsonapi.StateVersionCreateVersionOptions{
		Lineage: &state.Lineage,
		MD5:     otf.String(fmt.Sprintf("%x", md5.Sum(opts.State))),
		Serial:  otf.Int64(state.Serial),
		State:   otf.String(base64.StdEncoding.EncodeToString(opts.State)),
	})
	if err != nil {
		return nil, err
	}

	sv := jsonapi.StateVersion{}
	if err = c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}

	return &Version{ID: sv.ID, Serial: sv.Serial}, nil
}

func (c *Client) DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error) {
	// two steps:
	// 1) retrieve current state version for the workspace
	// 2) use the download link to download the state data
	u := fmt.Sprintf("workspaces/%s/current-state-version", url.QueryEscape(workspaceID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	v := &jsonapi.StateVersion{}
	err = c.Do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	req, err = c.NewRequest("GET", v.DownloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	var buf bytes.Buffer
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
