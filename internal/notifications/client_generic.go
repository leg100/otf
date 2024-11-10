package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
)

type (
	// GenericPayload is the information sent in generic notifications, as
	// documented here:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/notification-configurations#run-notification-payload
	GenericPayload struct {
		PayloadVersion              int
		NotificationConfigurationID resource.ID
		RunURL                      string
		RunID                       resource.ID
		RunMessage                  string
		RunCreatedAt                time.Time
		RunCreatedBy                string
		WorkspaceID                 resource.ID
		WorkspaceName               string
		OrganizationName            string
		Notifications               []genericNotificationPayload
	}

	genericNotificationPayload struct {
		Message      string
		Trigger      Trigger
		RunStatus    run.Status
		RunUpdatedAt time.Time
		RunUpdatedBy string
	}

	genericClient struct {
		client *http.Client
		url    string
	}
)

func newGenericClient(cfg *Config) (*genericClient, error) {
	return &genericClient{
		client: &http.Client{},
		url:    *cfg.URL,
	}, nil
}

func (c *genericClient) Publish(ctx context.Context, n *notification) error {
	payload, err := n.genericPayload()
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
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

func (c *genericClient) Close() {
	c.client.CloseIdleConnections()
}
