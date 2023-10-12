package releases

import (
	"bytes"
	"context"
	"time"

	"github.com/leg100/otf/internal"
)

type Client struct {
	internal.JSONAPIClient

	*downloader
}

func NewClient(client internal.JSONAPIClient, destdir string) *Client {
	c := &Client{
		JSONAPIClient: client,
	}
	c.downloader = newDownloader(destdir, c)
	return c
}

func (c *Client) getLatest(ctx context.Context) (string, time.Time, error) {
	req, err := c.NewRequest("GET", "releases/latest", nil)
	if err != nil {
		return "", time.Time{}, err
	}
	v := new(bytes.Buffer)
	if err = c.Do(ctx, req, v); err != nil {
		return "", time.Time{}, err
	}
	return v.String(), time.Time{}, nil
}
