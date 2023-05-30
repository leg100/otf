package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"

	"cloud.google.com/go/pubsub"
)

var (
	ErrInvalidGoogleProjectID    = errors.New("URL host must be a valid GCP project ID")
	ErrInvalidGooglePubSubTopic  = errors.New("URL path must be a valid GCP pubsub topic ID")
	ErrInvalidGooglePubSubScheme = errors.New("URL scheme must be: " + gcpPubSubScheme)

	_ client = (*pubsubClient)(nil)

	// URL scheme for destination type gcp pubsub, e.g.
	// gcppubsub://<project_id/<topic_name>
	gcpPubSubScheme = "gcppubsub"

	// regex for a google project ID:
	//
	// https://cloud.google.com/resource-manager/docs/creating-managing-projects#before_you_begin
	gcpProjectIDRegex = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]$`)
	// regex for gcp pubsub topic name:
	//
	// https://cloud.google.com/pubsub/docs/create-topic#resource_names
	gcpPubSubTopicRegex = regexp.MustCompile(`^[a-zA-Z][-a-zA-Z0-9]{2,254}$`)
)

type (
	pubsubClient struct {
		client *pubsub.Client
		topic  *pubsub.Topic
	}
)

func newPubSubClient(cfg *Config) (*pubsubClient, error) {
	if cfg.URL == nil {
		return nil, ErrDestinationRequiresURL
	}
	u, err := url.Parse(*cfg.URL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != gcpPubSubScheme {
		return nil, ErrInvalidGooglePubSubScheme
	}

	if !gcpProjectIDRegex.MatchString(u.Host) {
		return nil, ErrInvalidGoogleProjectID
	}
	project := u.Host

	if len(u.Path) == 0 || u.Path[0] != '/' || !gcpPubSubTopicRegex.MatchString(u.Path[1:]) {
		return nil, ErrInvalidGooglePubSubTopic
	}
	topic := u.Path[1:]

	client, err := pubsub.NewClient(context.Background(), project)
	if err != nil {
		return nil, err
	}
	return &pubsubClient{
		client: client,
		topic:  client.Topic(topic),
	}, nil
}

func (c *pubsubClient) Publish(ctx context.Context, n *notification) error {
	payload, err := n.genericPayload()
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	c.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	return nil
}

func (c *pubsubClient) Close() {
	c.client.Close()
}
