package notifications

import (
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

// TFENotificationTriggerType represents the different TFE notifications that can be sent
// as a run's progress transitions between different states
type TFENotificationTriggerType string

const (
	NotificationTriggerCreated               TFENotificationTriggerType = "run:created"
	NotificationTriggerPlanning              TFENotificationTriggerType = "run:planning"
	NotificationTriggerNeedsAttention        TFENotificationTriggerType = "run:needs_attention"
	NotificationTriggerApplying              TFENotificationTriggerType = "run:applying"
	NotificationTriggerCompleted             TFENotificationTriggerType = "run:completed"
	NotificationTriggerErrored               TFENotificationTriggerType = "run:errored"
	NotificationTriggerAssessmentDrifted     TFENotificationTriggerType = "assessment:drifted"
	NotificationTriggerAssessmentFailed      TFENotificationTriggerType = "assessment:failed"
	NotificationTriggerAssessmentCheckFailed TFENotificationTriggerType = "assessment:check_failure"
)

// TFENotificationDestinationType represents the destination type of the
// notification configuration.
type TFENotificationDestinationType string

// List of available notification destination types.
const (
	NotificationDestinationTypeEmail          TFENotificationDestinationType = "email"
	NotificationDestinationTypeGeneric        TFENotificationDestinationType = "generic"
	NotificationDestinationTypeSlack          TFENotificationDestinationType = "slack"
	NotificationDestinationTypeMicrosoftTeams TFENotificationDestinationType = "microsoft-teams"
)

func TFENotificationDestinationPtr(d TFENotificationDestinationType) *TFENotificationDestinationType {
	return &d
}

// TFENotificationConfigurationList represents a list of Notification
// Configurations.
type TFENotificationConfigurationList struct {
	*types.Pagination
	Items []*TFENotificationConfiguration
}

// TFENotificationConfiguration represents a Notification Configuration.
type TFENotificationConfiguration struct {
	ID                resource.TfeID                 `jsonapi:"primary,notification-configurations"`
	CreatedAt         time.Time                      `jsonapi:"attribute" json:"created-at"`
	DeliveryResponses []*TFEDeliveryResponse         `jsonapi:"attribute" json:"delivery-responses"`
	DestinationType   TFENotificationDestinationType `jsonapi:"attribute" json:"destination-type"`
	Enabled           bool                           `jsonapi:"attribute" json:"enabled"`
	Name              string                         `jsonapi:"attribute" json:"name"`
	Token             string                         `jsonapi:"attribute" json:"token"`
	Triggers          []string                       `jsonapi:"attribute" json:"triggers"`
	UpdatedAt         time.Time                      `jsonapi:"attribute" json:"updated-at"`
	URL               string                         `jsonapi:"attribute" json:"url"`

	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses"`
	// relationships
	Subscribable *workspace.TFEWorkspace `jsonapi:"relationship" json:"subscribable"`
	EmailUsers   []*user.TFEUser         `jsonapi:"relationship" json:"users"`
}

// TFEDeliveryResponse represents a notification configuration delivery response.
type TFEDeliveryResponse struct {
	Body       string              `jsonapi:"attribute" json:"body"`
	Code       string              `jsonapi:"attribute" json:"code"`
	Headers    map[string][]string `jsonapi:"attribute" json:"headers"`
	SentAt     time.Time           `jsonapi:"attribute" json:"sent-at"`
	Successful string              `jsonapi:"attribute" json:"successful"`
	URL        string              `jsonapi:"attribute" json:"url"`
}

// TFENotificationConfigurationCreateOptions represents the options for
// creating a new notification configuration.
type TFENotificationConfigurationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,notification-configurations"`

	// Required: The destination type of the notification configuration
	DestinationType *TFENotificationDestinationType `jsonapi:"attribute" json:"destination-type"`

	// Required: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attribute" json:"enabled"`

	// Required: The name of the notification configuration
	Name *string `jsonapi:"attribute" json:"name"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attribute" json:"token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []TFENotificationTriggerType `jsonapi:"attribute" json:"triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attribute" json:"url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*user.TFEUser `jsonapi:"relationship" json:"users,omitempty"`
}

// TFENotificationConfigurationUpdateOptions represents the options for
// updating a existing notification configuration.
type TFENotificationConfigurationUpdateOptions struct {
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
	Triggers []TFENotificationTriggerType `jsonapi:"attribute" json:"triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attribute" json:"url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attribute" json:"email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*user.TFEUser `jsonapi:"relationship" json:"users,omitempty"`
}
