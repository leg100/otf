package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var _ client = (*slackClient)(nil)

type (
	slackClient struct {
		*genericClient
	}
	slackMessage struct {
		Text string `json:"text"`
	}
)

func newSlackClient(cfg *Config) (*slackClient, error) {
	client, err := newGenericClient(cfg)
	if err != nil {
		return nil, err
	}
	return &slackClient{
		genericClient: client,
	}, nil
}

func (c *slackClient) Publish(ctx context.Context, n *notification) error {
	text := fmt.Sprintf("new run update: %s: %s", n.run.ID, n.run.Status)
	data, err := json.Marshal(slackMessage{Text: text})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
