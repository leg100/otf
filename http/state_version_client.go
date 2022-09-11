package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

// Create a new state version for the given workspace.
func (c *client) CreateStateVersion(ctx context.Context, workspaceID string, opts otf.StateVersionCreateOptions) (*otf.StateVersion, error) {
	u := fmt.Sprintf("workspaces/%s/state-versions", url.QueryEscape(workspaceID))
	req, err := c.newRequest("POST", u, &dto.StateVersionCreateOptions{
		Lineage: opts.Lineage,
		MD5:     opts.MD5,
		Serial:  opts.Serial,
		State:   opts.State,
		Run:     &dto.Run{ID: opts.Run.ID()},
	})
	if err != nil {
		return nil, err
	}

	sv := &dto.StateVersion{}
	err = c.do(ctx, req, sv)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalStateVersionJSONAPI(sv), nil
}

func (c *client) CurrentStateVersion(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
	u := fmt.Sprintf("workspaces/%s/current-state-version", url.QueryEscape(workspaceID))
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	sv := &dto.StateVersion{}
	err = c.do(ctx, req, sv)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalStateVersionJSONAPI(sv), nil
}

// DownloadStateVersion retrieves the actual stored state of a state version
func (c *client) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	u := fmt.Sprintf("state-versions/%s/download", url.QueryEscape(svID))
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	var buf bytes.Buffer
	err = c.do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
