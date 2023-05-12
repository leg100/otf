package notifications

import (
	"github.com/leg100/otf/internal/run"
)

type client interface {
	Publish(run *run.Run) error
	Close()
}

func newClient(cfg *Config) (client, error) {
	switch cfg.DestinationType {
	case NotificationDestinationTypeSlack:
		return newSlackClient(cfg)
	default:
		return nil, ErrUnsupportedDestinationType
	}
}
