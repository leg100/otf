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
		client *http.Client
		url    string
	}
	slackMessage struct {
		Text string `json:"text"`
	}
)

func newSlackClient(cfg *Config) (*slackClient, error) {
	return &slackClient{
		client: &http.Client{},
		url:    *cfg.URL,
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

func (c *slackClient) Close() {
	c.client.CloseIdleConnections()
}
