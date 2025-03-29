package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// NotificationTriggerType represents the different TFE notifications that can be sent
// as a run's progress transitions between different states
type NotificationTriggerType string

const (
	NotificationTriggerCreated               NotificationTriggerType = "run:created"
	NotificationTriggerPlanning              NotificationTriggerType = "run:planning"
	NotificationTriggerNeedsAttention        NotificationTriggerType = "run:needs_attention"
	NotificationTriggerApplying              NotificationTriggerType = "run:applying"
	NotificationTriggerCompleted             NotificationTriggerType = "run:completed"
	NotificationTriggerErrored               NotificationTriggerType = "run:errored"
	NotificationTriggerAssessmentDrifted     NotificationTriggerType = "assessment:drifted"
	NotificationTriggerAssessmentFailed      NotificationTriggerType = "assessment:failed"
	NotificationTriggerAssessmentCheckFailed NotificationTriggerType = "assessment:check_failure"
)

// NotificationDestinationType represents the destination type of the
// notification configuration.
type NotificationDestinationType string

// List of available notification destination types.
const (
	NotificationDestinationTypeEmail          NotificationDestinationType = "email"
	NotificationDestinationTypeGeneric        NotificationDestinationType = "generic"
	NotificationDestinationTypeSlack          NotificationDestinationType = "slack"
	NotificationDestinationTypeMicrosoftTeams NotificationDestinationType = "microsoft-teams"
)

func NotificationDestinationPtr(d NotificationDestinationType) *NotificationDestinationType {
	return &d
}

// NotificationConfigurationList represents a list of Notification
// Configurations.
type NotificationConfigurationList struct {
	*Pagination
	Items []*NotificationConfiguration
}

// NotificationConfiguration represents a Notification Configuration.
type NotificationConfiguration struct {
	ID                resource.ID                 `jsonapi:"primary,notification-configurations"`
	CreatedAt         time.Time                   `jsonapi:"attribute" json:"created-at"`
	DeliveryResponses []*DeliveryResponse         `jsonapi:"attribute" json:"delivery-responses"`
	DestinationType   NotificationDestinationType `jsonapi:"attribute" json:"destination-type"`
	Enabled           bool                        `jsonapi:"attribute" json:"enabled"`
	Name              string                      `jsonapi:"attribute" json:"name"`
	Token             string                      `jsonapi:"attribute" json:"token"`
	Triggers          []string                    `jsonapi:"attribute" json:"triggers"`
	UpdatedAt         time.Time                   `jsonapi:"attribute" json:"updated-at"`
	URL               string                      `jsonapi:"attribute" json:"url"`

	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses"`
	// relationships
	Subscribable *Workspace `jsonapi:"relationship" json:"subscribable"`
	EmailUsers   []*User    `jsonapi:"relationship" json:"users"`
}

// DeliveryResponse represents a notification configuration delivery response.
type DeliveryResponse struct {
	Body       string              `jsonapi:"attribute" json:"body"`
	Code       string              `jsonapi:"attribute" json:"code"`
	Headers    map[string][]string `jsonapi:"attribute" json:"headers"`
	SentAt     time.Time           `jsonapi:"attribute" json:"sent-at"`
	Successful string              `jsonapi:"attribute" json:"successful"`
	URL        string              `jsonapi:"attribute" json:"url"`
}

// NotificationConfigurationCreateOptions represents the options for
// creating a new notification configuration.
type NotificationConfigurationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,notification-configurations"`

	// Required: The destination type of the notification configuration
	DestinationType *NotificationDestinationType `jsonapi:"attribute" json:"destination-type"`

	// Required: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attribute" json:"enabled"`

	// Required: The name of the notification configuration
	Name *string `jsonapi:"attribute" json:"name"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attribute" json:"token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attribute" json:"triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attribute" json:"url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relationship" json:"users,omitempty"`
}

// NotificationConfigurationUpdateOptions represents the options for
// updating a existing notification configuration.
type NotificationConfigurationUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,notification-configurations"`

	// Optional: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attribute" json:"enabled,omitempty"`

	// Optional: The name of the notification configuration
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attribute" json:"token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attribute" json:"triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attribute" json:"url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relationship" json:"users,omitempty"`
}
