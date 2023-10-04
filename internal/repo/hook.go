package repo

import (
	"log/slog"
	"path"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

// defaultEvents are the VCS events that hooks subscribe to.
var defaultEvents = []vcs.EventType{
	vcs.EventTypePush,
	vcs.EventTypePull,
}

type (
	// hook is a webhook for a VCS repo
	hook struct {
		id            uuid.UUID // internal otf ID
		cloudID       *string   // cloud's hook ID; populated following synchronisation
		vcsProviderID string

		secret     string   // secret token
		identifier string   // repo identifier: <repo_owner>/<repo_name>
		cloud      vcs.Kind // origin of events
		endpoint   string   // OTF URL that receives events
	}

	newHookOptions struct {
		id            *uuid.UUID
		vcsProviderID string
		secret        *string
		identifier    string
		cloud         vcs.Kind
		cloudID       *string // cloud's webhook id

		// for building endpoint URL
		internal.HostnameService
	}
)

func newHook(opts newHookOptions) (*hook, error) {
	hook := hook{
		identifier:    opts.identifier,
		cloud:         opts.cloud,
		cloudID:       opts.cloudID,
		vcsProviderID: opts.vcsProviderID,
	}

	if opts.id != nil {
		hook.id = *opts.id
	} else {
		hook.id = uuid.New()
	}

	if opts.secret != nil {
		hook.secret = *opts.secret
	} else {
		secret, err := internal.GenerateToken()
		if err != nil {
			return nil, err
		}
		hook.secret = secret
	}

	hook.endpoint = opts.URL(path.Join(handlerPrefix, hook.id.String()))

	return &hook, nil
}

func (h *hook) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", h.id.String()),
		slog.String("vcs_provider_id", h.vcsProviderID),
		slog.String("cloud", string(h.cloud)),
		slog.String("repo", h.identifier),
		slog.String("endpoint", h.endpoint),
	}
	if h.cloudID != nil {
		attrs = append(attrs, slog.String("vcs_id", *h.cloudID))
	}
	return slog.GroupValue(attrs...)
}
