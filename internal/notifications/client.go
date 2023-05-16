package notifications

import "context"

type (
	// client is a client capable of sending notifications to third party
	client interface {
		// Publish notification. The run and workspace relating to the event are
		// provided with which to populate the notification.
		Publish(ctx context.Context, n *notification) error
		// Close the client to free up resources.
		Close()
	}

	clientFactory interface {
		newClient(*Config) (client, error)
	}

	defaultFactory struct{}
)

func (f *defaultFactory) newClient(cfg *Config) (client, error) {
	switch cfg.DestinationType {
	case DestinationGeneric:
		return newGenericClient(cfg)
	case DestinationSlack:
		return newSlackClient(cfg)
	case DestinationGCPPubSub:
		return newPubSubClient(cfg)
	default:
		return nil, ErrUnsupportedDestination
	}
}
