package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal/run"
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

func (c *slackClient) Publish(r *run.Run) error {
	text := fmt.Sprintf("new run update: %s: %s", r.ID, r.Status)
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
