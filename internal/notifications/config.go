package notifications

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
)

const (
	DestinationGeneric   Destination = "generic"
	DestinationSlack     Destination = "slack"
	DestinationGCPPubSub Destination = "gcppubsub"
	// Email type is only accepted in order to pass the `go-tfe` API tests,
	// which create configs with this type. It otherwise is entirely
	// unfunctional; no emails are sent.
	DestinationEmail Destination = "email"

	TriggerCreated        Trigger = "run:created"
	TriggerPlanning       Trigger = "run:planning"
	TriggerNeedsAttention Trigger = "run:needs_attention"
	TriggerApplying       Trigger = "run:applying"
	TriggerCompleted      Trigger = "run:completed"
	TriggerErrored        Trigger = "run:errored"
)

var (
	ErrUnsupportedDestination = errors.New("unsupported notification destination")
	ErrDestinationRequiresURL = errors.New("URL must be specified for this destination")
	ErrInvalidTrigger         = errors.New("invalid notification trigger")
)

type (
	// Config represents a Notification Configuration.
	Config struct {
		ID              resource.TfeID `json:"notification_configuration_id" db:"notification_configuration_id"`
		CreatedAt       time.Time      `json:"created_at" db:"created_at"`
		UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
		DestinationType Destination    `json:"destination_type" db:"destination_type"`
		Enabled         bool
		Name            string
		Triggers        []Trigger
		URL             *string
		WorkspaceID     resource.TfeID `json:"workspace_id" db:"workspace_id"`
	}

	// Trigger is the event triggering a notification
	Trigger string

	// Destination is the destination platform for an event.
	Destination string

	CreateConfigOptions struct {
		// Required: The destination type of the notification configuration
		DestinationType Destination

		// Required: Whether the notification configuration should be enabled or not
		Enabled *bool

		// Required: The name of the notification configuration
		Name *string

		// Optional: The token of the notification configuration
		Token *string

		// Optional: The list of run events that will trigger notifications.
		Triggers []Trigger

		// Optional: The url of the notification configuration
		URL *string
	}

	// UpdateConfigOptions represents the options for
	// updating a existing notification configuration.
	UpdateConfigOptions struct {
		// Optional: Whether the notification configuration should be enabled or not
		Enabled *bool

		// Optional: The name of the notification configuration
		Name *string

		// Optional: The token of the notification configuration
		Token *string

		// Optional: The list of run events that will trigger notifications.
		Triggers []Trigger

		// Optional: The url of the notification configuration
		URL *string
	}
)

func NewConfig(workspaceID resource.TfeID, opts CreateConfigOptions) (*Config, error) {
	if opts.DestinationType != DestinationGeneric &&
		opts.DestinationType != DestinationEmail &&
		opts.DestinationType != DestinationSlack &&
		opts.DestinationType != DestinationGCPPubSub {
		return nil, ErrUnsupportedDestination
	}
	// an empty url is only acceptable with the email type
	if opts.DestinationType != DestinationEmail &&
		opts.URL == nil {
		return nil, fmt.Errorf("must specify url for this destination type")
	}
	// if a url is supplied it must be valid
	if opts.URL != nil {
		if _, err := url.Parse(*opts.URL); err != nil {
			return nil, fmt.Errorf("invalid url: %w", err)
		}
	}
	if err := validTriggers(opts.Triggers); err != nil {
		return nil, err
	}
	if opts.Enabled == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "enabled"}
	}
	if opts.Name == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "name"}
	}
	if *opts.Name == "" {
		return nil, fmt.Errorf("name cannot be an empty string")
	}

	return &Config{
		ID:              resource.NewTfeID(resource.NotificationConfigurationKind),
		CreatedAt:       internal.CurrentTimestamp(nil),
		UpdatedAt:       internal.CurrentTimestamp(nil),
		Name:            *opts.Name,
		Enabled:         *opts.Enabled,
		Triggers:        opts.Triggers,
		DestinationType: opts.DestinationType,
		URL:             opts.URL,
		WorkspaceID:     workspaceID,
	}, nil
}

func (c *Config) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("name", c.Name),
		slog.Bool("enabled", c.Enabled),
		slog.Any("triggers", c.Triggers),
		slog.String("workspace_id", c.WorkspaceID.String()),
		slog.String("destination", string(c.DestinationType)),
	}
	return slog.GroupValue(attrs...)
}

func (c *Config) update(opts UpdateConfigOptions) error {
	if opts.Name != nil {
		if *opts.Name == "" {
			return fmt.Errorf("name cannot be an empty string")
		}
		c.Name = *opts.Name
	}
	if opts.Enabled != nil {
		c.Enabled = *opts.Enabled
	}
	if err := validTriggers(opts.Triggers); err != nil {
		return err
	}
	if opts.Triggers != nil {
		c.Triggers = opts.Triggers
	}
	if opts.URL != nil {
		c.URL = opts.URL
	}
	return nil
}

// matchTrigger determines whether the config has a trigger that matches the
// given run status
func (c *Config) matchTrigger(status runstatus.Status) (Trigger, bool) {
	switch status {
	case runstatus.Pending:
		return TriggerCreated, c.hasTrigger(TriggerCreated)
	case runstatus.Planning:
		return TriggerPlanning, c.hasTrigger(TriggerPlanning)
	case runstatus.Planned:
		return TriggerNeedsAttention, c.hasTrigger(TriggerNeedsAttention)
	case runstatus.Applying:
		return TriggerApplying, c.hasTrigger(TriggerApplying)
	case runstatus.Errored:
		return TriggerErrored, c.hasTrigger(TriggerErrored)
	}
	if runstatus.Done(status) {
		return TriggerCompleted, c.hasTrigger(TriggerCompleted)
	}
	return "", false
}

func (c *Config) hasTrigger(t Trigger) bool {
	return slices.Contains(c.Triggers, t)
}

func validTriggers(triggers []Trigger) error {
	for _, t := range triggers {
		switch t {
		case TriggerCreated,
			TriggerPlanning,
			TriggerNeedsAttention,
			TriggerApplying,
			TriggerCompleted,
			TriggerErrored:
		default:
			return ErrInvalidTrigger
		}
	}
	return nil
}
