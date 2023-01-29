package state

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
)

// jsonapiClient is a state version client using json-api
type jsonapiClient struct {
	otf.JSONAPIClient
}

// Create a new state version for the given workspace.
func (c *jsonapiClient) CreateStateVersion(ctx context.Context, workspaceID string, opts versionJSONAPICreateOptions) (*StateVersion, error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(workspaceID))
	req, err := c.NewRequest("POST", u, &versionJSONAPICreateOptions{
		Lineage: opts.Lineage,
		MD5:     opts.MD5,
		Serial:  opts.Serial,
		State:   opts.State,
	})
	if err != nil {
		return nil, err
	}

	sv := &versionJSONAPI{}
	err = c.Do(ctx, req, sv)
	if err != nil {
		return nil, err
	}

	return unmarshalJSONAPI(sv), nil
}

func (c *jsonapiClient) CurrentStateVersion(ctx context.Context, workspaceID string) (*StateVersion, error) {
	u := fmt.Sprintf("workspaces/%s/current-state-version", url.QueryEscape(workspaceID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	sv := &versionJSONAPI{}
	err = c.Do(ctx, req, sv)
	if err != nil {
		return nil, err
	}

	return unmarshalJSONAPI(sv), nil
}

// DownloadStateVersion retrieves the actual stored state of a state version
func (c *jsonapiClient) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	u := fmt.Sprintf("state-versions/%s/download", url.QueryEscape(svID))
	req, err := c.NewRequest("GET", u, nil)
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
