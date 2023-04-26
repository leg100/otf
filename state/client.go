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
	var state File
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

	return newFromJSONAPI(&sv), nil
}

func (c *Client) ListStateVersions(ctx context.Context, options StateVersionListOptions) (*VersionList, error) {
	req, err := c.NewRequest("GET", "state-versions", &options)
	if err != nil {
		return nil, err
	}

	list := &jsonapi.StateVersionList{}
	err = c.Do(ctx, req, list)
	if err != nil {
		return nil, err
	}

	return newListFromJSONAPI(list), nil
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

	sv := jsonapi.StateVersion{}
	if err := c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}
	return newFromJSONAPI(&sv), nil
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
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape("ws-rollback"))
	req, err := c.NewRequest("PATCH", u, &jsonapi.RollbackStateVersionOptions{
		RollbackStateVersion: &jsonapi.StateVersion{ID: svID},
	})
	if err != nil {
		return nil, err
	}

	sv := jsonapi.StateVersion{}
	if err = c.Do(ctx, req, &sv); err != nil {
		return nil, err
	}

	return newFromJSONAPI(&sv), nil
}

func newFromJSONAPI(from *jsonapi.StateVersion) *Version {
	return &Version{
		ID:     from.ID,
		Serial: from.Serial,
	}
}

// newListFromJSONAPI constructs a state version list from a json:api struct
func newListFromJSONAPI(from *jsonapi.StateVersionList) *VersionList {
	to := VersionList{
		Pagination: jsonapi.NewPaginationFromJSONAPI(from.Pagination),
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, newFromJSONAPI(i))
	}
	return &to
}
