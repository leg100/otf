package http

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/leg100/otf"
)

func (c *client) PutChunk(ctx context.Context, chunk otf.Chunk) error {
	u := fmt.Sprintf("runs/%s/logs/%s", url.QueryEscape(chunk.RunID), url.QueryEscape(string(chunk.Phase)))
	req, err := c.newRequest("PUT", u, chunk.Data)
	if err != nil {
		return err
	}
	// newRequest() only lets us set a query or a payload but not both, so we
	// set query here.
	q := url.Values{}
	q.Add("offset", strconv.Itoa(chunk.Offset))
	req.URL.RawQuery = q.Encode()

	err = c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
