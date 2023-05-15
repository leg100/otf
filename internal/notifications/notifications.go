// Package notifications sends notifications for run state transitions and
// workspace events.
package notifications

import (
	"net/url"

	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

// notification furnishes information for sending a notification to a third
// party.
type notification struct {
	workspace *workspace.Workspace
	run       *run.Run
	trigger   Trigger
	config    *Config
	hostname  string
}

// genericPayload converts a notification into a format suitable for the generic
// and GCP-pubsub destination types.
func (n *notification) genericPayload() (*GenericPayload, error) {
	runUpdatedAt, err := n.run.StatusTimestamp(n.run.Status)
	if err != nil {
		return nil, err
	}
	return &GenericPayload{
		PayloadVersion:              1,
		NotificationConfigurationID: "",
		RunURL:                      n.runURL(),
		RunID:                       n.run.ID,
		RunCreatedAt:                n.run.CreatedAt,
		WorkspaceID:                 n.workspace.ID,
		WorkspaceName:               n.workspace.Name,
		OrganizationName:            n.workspace.Organization,
		Notifications: []genericNotificationPayload{
			{
				Trigger:      n.trigger,
				RunStatus:    n.run.Status,
				RunUpdatedAt: runUpdatedAt,
			},
		},
	}, nil
}

func (n *notification) runURL() string {
	u := &url.URL{Scheme: "https", Host: n.hostname, Path: paths.Run(n.run.ID)}
	return u.String()
}
