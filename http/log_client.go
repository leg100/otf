package http

import (
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
)

func (c *client) PutChunk(ctx context.Context, runID string, phase otf.PhaseType, chunk otf.Chunk) error {
	u := fmt.Sprintf("runs/%s/logs/%s", url.QueryEscape(runID), url.QueryEscape(string(phase)))
	req, err := c.newRequest("PUT", u, chunk.Marshal())
	if err != nil {
		return err
	}

	err = c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
