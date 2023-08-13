package repo

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/cloud"
)

// defaultEvents are the VCS events that hooks subscribe to.
var defaultEvents = []cloud.VCSEventType{
	cloud.VCSEventTypePush,
	cloud.VCSEventTypePull,
}

// hook is a webhook for a VCS repo
type hook struct {
	id            uuid.UUID // internal otf ID
	cloudID       *string   // cloud's hook ID; populated following synchronisation
	vcsProviderID string

	secret     string // secret token
	identifier string // repo identifier: <repo_owner>/<repo_name>
	cloud      string // cloud name
	endpoint   string // otf URL that receives events

	cloud.EventHandler // handles incoming vcs events
}

func (h *hook) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", h.id.String()),
		slog.String("vcs_provider_id", h.vcsProviderID),
		slog.String("cloud", h.cloud),
		slog.String("repo", h.identifier),
		slog.String("endpoint", h.endpoint),
	}
	if h.cloudID != nil {
		attrs = append(attrs, slog.String("vcs_id", *h.cloudID))
	}
	return slog.GroupValue(attrs...)
}
