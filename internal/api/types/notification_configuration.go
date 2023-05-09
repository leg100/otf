package types

import "time"

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

// NotificationConfigurationList represents a list of Notification
// Configurations.
type NotificationConfigurationList struct {
	*Pagination
	Items []*NotificationConfiguration
}

// NotificationConfiguration represents a Notification Configuration.
type NotificationConfiguration struct {
	ID                string                      `jsonapi:"primary,notification-configurations"`
	CreatedAt         time.Time                   `jsonapi:"attr,created-at,iso8601"`
	DeliveryResponses []*DeliveryResponse         `jsonapi:"attr,delivery-responses"`
	DestinationType   NotificationDestinationType `jsonapi:"attr,destination-type"`
	Enabled           bool                        `jsonapi:"attr,enabled"`
	Name              string                      `jsonapi:"attr,name"`
	Token             string                      `jsonapi:"attr,token"`
	Triggers          []string                    `jsonapi:"attr,triggers"`
	UpdatedAt         time.Time                   `jsonapi:"attr,updated-at,iso8601"`
	URL               string                      `jsonapi:"attr,url"`

	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses"`

	// Relations
	Subscribable *Workspace `jsonapi:"relation,subscribable"`
	EmailUsers   []*User    `jsonapi:"relation,users"`
}

// DeliveryResponse represents a notification configuration delivery response.
type DeliveryResponse struct {
	Body       string              `jsonapi:"attr,body"`
	Code       string              `jsonapi:"attr,code"`
	Headers    map[string][]string `jsonapi:"attr,headers"`
	SentAt     time.Time           `jsonapi:"attr,sent-at,rfc3339"`
	Successful string              `jsonapi:"attr,successful"`
	URL        string              `jsonapi:"attr,url"`
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
	DestinationType *NotificationDestinationType `jsonapi:"attr,destination-type"`

	// Required: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attr,enabled"`

	// Required: The name of the notification configuration
	Name *string `jsonapi:"attr,name"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attr,token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attr,triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attr,url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relation,users,omitempty"`
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
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`

	// Optional: The name of the notification configuration
	Name *string `jsonapi:"attr,name,omitempty"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attr,token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attr,triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attr,url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relation,users,omitempty"`
}
