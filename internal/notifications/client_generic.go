package notifications

import (
	"time"

	"github.com/leg100/otf/internal"
)

type (
	// GenericPayload is the information sent in generic notifications, as
	// documented here:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/notification-configurations#run-notification-payload
	GenericPayload struct {
		PayloadVersion              int
		NotificationConfigurationID string
		RunURL                      string
		RunID                       string
		RunMessage                  string
		RunCreatedAt                time.Time
		RunCreatedBy                string
		WorkspaceID                 string
		WorkspaceName               string
		OrganizationName            string
		Notifications               []genericNotificationPayload
	}

	genericNotificationPayload struct {
		Message      string
		Trigger      Trigger
		RunStatus    internal.RunStatus
		RunUpdatedAt time.Time
		RunUpdatedBy string
	}
)
