package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var _ client = (*slackClient)(nil)

type (
	slackClient struct {
		*genericClient
	}
	slackMessage struct {
		Blocks []slackBlock `json:"blocks"`
	}
	slackBlock struct {
		Type string `json:"type"`
		Text any    `json:"text"`
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
	data, err := json.Marshal(slackMessage{
		Blocks: []slackBlock{
			{
				Type: "section",
				Text: &slackBlock{
					Type: "mrkdwn",
					Text: fmt.Sprintf("Run notification for <%s|%s/%s>", n.runURL(), n.workspace.Organization, n.workspace.Name),
				},
			},
			{
				Type: "section",
				Text: &slackBlock{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*run %s*", strings.ReplaceAll(string(n.run.Status), "_", " ")),
				},
			},
		},
	})
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
