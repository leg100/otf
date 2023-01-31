package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
)

// Download a configuration version.  Only configuration versions in the uploaded state may be downloaded.
func (c *client) DownloadConfig(ctx context.Context, cvID string) ([]byte, error) {
	u := fmt.Sprintf("configuration-versions/%s/download", url.QueryEscape(cvID))
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
