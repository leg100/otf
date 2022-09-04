package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (c *client) GetStateVersion(ctx context.Context, workspaceID string) (*otf.StateVersion, error) {
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
func (c *client) DownloadStateVersion(ctx context.Context, u string) ([]byte, error) {
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
