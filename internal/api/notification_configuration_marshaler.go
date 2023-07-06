package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/notifications"
)

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (m *jsonapiMarshaler) toNotificationConfig(from *notifications.Config) *types.NotificationConfiguration {
	to := &types.NotificationConfiguration{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		UpdatedAt:       from.UpdatedAt,
		Name:            from.Name,
		Enabled:         from.Enabled,
		DestinationType: types.NotificationDestinationType(from.DestinationType),
		Subscribable: &types.Workspace{
			ID: from.WorkspaceID,
		},
	}
	if from.URL != nil {
		to.URL = *from.URL
	}
	for _, t := range from.Triggers {
		to.Triggers = append(to.Triggers, string(t))
	}
	return to
}
