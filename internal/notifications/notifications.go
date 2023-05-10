// Package notifications sends notifications for run state transitions and
// workspace events.
package notifications

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"golang.org/x/exp/slog"
)

const (
	NotificationDestinationTypeGeneric   Destination = "generic"
	NotificationDestinationTypeGCPPubSub Destination = "gcppubsub"

	NotificationTriggerCreated               Trigger = "run:created"
	NotificationTriggerPlanning              Trigger = "run:planning"
	NotificationTriggerNeedsAttention        Trigger = "run:needs_attention"
	NotificationTriggerApplying              Trigger = "run:applying"
	NotificationTriggerCompleted             Trigger = "run:completed"
	NotificationTriggerErrored               Trigger = "run:errored"
	NotificationTriggerAssessmentDrifted     Trigger = "assessment:drifted"
	NotificationTriggerAssessmentFailed      Trigger = "assessment:failed"
	NotificationTriggerAssessmentCheckFailed Trigger = "assessment:check_failure"
)

var (
	ErrUnsupportedDestinationType = errors.New("unsupported destination type")
	ErrInvalidTrigger             = errors.New("invalid notification trigger")
)

type (
	// Config represents a Notification Configuration.
	Config struct {
		ID              string
		CreatedAt       time.Time
		UpdatedAt       time.Time
		DestinationType Destination
		Enabled         bool
		Name            string
		Token           string
		Triggers        []Trigger
		URL             *string
		WorkspaceID     string
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

func NewConfig(workspaceID string, opts CreateConfigOptions) (*Config, error) {
	if err := validDestinationType(opts.DestinationType); err != nil {
		return nil, err
	}
	if err := validTriggers(opts.Triggers); err != nil {
		return nil, err
	}
	if opts.Enabled == nil {
		return nil, &internal.MissingParameterError{Parameter: "enabled"}
	}
	if opts.Name == nil {
		return nil, &internal.MissingParameterError{Parameter: "name"}
	}
	if *opts.Name == "" {
		return nil, fmt.Errorf("name cannot be an empty string")
	}

	return &Config{
		ID:              internal.NewID("nc"),
		CreatedAt:       internal.CurrentTimestamp(),
		UpdatedAt:       internal.CurrentTimestamp(),
		Name:            *opts.Name,
		Enabled:         *opts.Enabled,
		Triggers:        opts.Triggers,
		DestinationType: opts.DestinationType,
		URL:             opts.URL,
		WorkspaceID:     workspaceID,
	}, nil
}

func validDestinationType(dt Destination) error {
	if dt != NotificationDestinationTypeGeneric &&
		dt != NotificationDestinationTypeGCPPubSub {
		return ErrUnsupportedDestinationType
	}
	return nil
}

func validTriggers(triggers []Trigger) error {
	for _, t := range triggers {
		switch t {
		case NotificationTriggerCreated,
			NotificationTriggerPlanning,
			NotificationTriggerNeedsAttention,
			NotificationTriggerApplying,
			NotificationTriggerCompleted,
			NotificationTriggerErrored,
			NotificationTriggerAssessmentDrifted,
			NotificationTriggerAssessmentFailed,
			NotificationTriggerAssessmentCheckFailed:
		default:
			return ErrInvalidTrigger
		}
	}
	return nil
}

func (c *Config) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("name", c.Name),
		slog.Bool("enabled", c.Enabled),
		slog.Any("triggers", c.Triggers),
		slog.String("workspace_id", c.WorkspaceID),
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
