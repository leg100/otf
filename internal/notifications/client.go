package notifications

import (
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// client is a client capable of sending notifications to third party
	client interface {
		// Publish notification. The run and workspace relating to the event are
		// provided with which to populate the notification.
		Publish(run *run.Run, ws *workspace.Workspace) error
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
	case DestinationSlack:
		return newSlackClient(cfg)
	case DestinationGCPPubSub:
		return newSlackClient(cfg)
	default:
		return nil, ErrUnsupportedDestination
	}
}
