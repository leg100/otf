package releases

import (
	"bytes"
	"context"
	"io"

	"github.com/leg100/otf/internal"
)

type Client struct {
	internal.JSONAPIClient

	*downloader
}

func NewClient(client internal.JSONAPIClient, destdir string) *Client {
	return &Client{
		JSONAPIClient: client,
		downloader:    newDownloader(destdir),
	}
}

// Download specified version of terraform. A network request is only made if
// "latest" is specified because only otfd knows the latest version.
func (c *Client) Download(ctx context.Context, version string, logger io.Writer) (string, error) {
	if version != "latest" {
		return c.downloader.Download(ctx, version, logger)
	}

	req, err := c.NewRequest("GET", "releases/latest", nil)
	if err != nil {
		return "", err
	}
	v := new(bytes.Buffer)
	if err = c.Do(ctx, req, v); err != nil {
		return "", err
	}
	return v.String(), nil
}
