package configversion

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfapi.Client

	// Client does not implement all of service yet
	Service
}

// DownloadConfig downloads a configuration version tarball.  Only configuration versions in the uploaded state may be downloaded.
func (c *Client) DownloadConfig(ctx context.Context, cvID resource.TfeID) ([]byte, error) {
	u := fmt.Sprintf("configuration-versions/%s/download", url.QueryEscape(cvID.String()))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
