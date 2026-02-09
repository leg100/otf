// Package notifications sends notifications for run state transitions and
// workspace events.
package notifications

import (
	"log/slog"
	"net/url"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/workspace"
)

// notification furnishes information for sending a notification to a third
// party.
type notification struct {
	event     pubsub.Event[*run.Event]
	workspace *workspace.Workspace
	trigger   Trigger
	config    *Config
	hostname  string
}

func (n *notification) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Time("time", n.event.Time),
		slog.String("workspace_id", n.workspace.ID.String()),
		slog.String("trigger", string(n.trigger)),
		slog.String("destination", string(n.config.DestinationType)),
	}
	return slog.GroupValue(attrs...)
}

// genericPayload converts a notification into a format suitable for the generic
// and GCP-pubsub destination types.
func (n *notification) genericPayload() (*GenericPayload, error) {
	return &GenericPayload{
		PayloadVersion:              1,
		NotificationConfigurationID: n.config.ID,
		RunURL:                      n.runURL(),
		RunID:                       n.event.Payload.ID,
		RunCreatedAt:                n.event.Payload.CreatedAt,
		WorkspaceID:                 n.workspace.ID,
		WorkspaceName:               n.workspace.Name,
		OrganizationName:            n.workspace.Organization,
		Notifications: []genericNotificationPayload{
			{
				Trigger:      n.trigger,
				RunStatus:    n.event.Payload.Status,
				RunUpdatedAt: n.event.Time,
			},
		},
	}, nil
}

func (n *notification) runURL() string {
	u := &url.URL{Scheme: "https", Host: n.hostname, Path: paths.Run(n.event.Payload.ID)}
	return u.String()
}
