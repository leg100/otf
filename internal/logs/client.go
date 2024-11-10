package logs

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	otfapi "github.com/leg100/otf/internal/api"
)

type Client struct {
	*otfapi.Client
}

func (c *Client) PutChunk(ctx context.Context, opts PutChunkOptions) error {
	u := fmt.Sprintf("runs/%s/logs/%s", url.QueryEscape(opts.RunID.String()), url.QueryEscape(string(opts.Phase)))
	req, err := c.NewRequest("PUT", u, opts.Data)
	if err != nil {
		return err
	}
	// newRequest() only lets us set a query or a payload but not both, so we
	// set query here.
	q := url.Values{}
	q.Add("offset", strconv.Itoa(opts.Offset))
	req.URL.RawQuery = q.Encode()

	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
