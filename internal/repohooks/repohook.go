// Package repohooks manages webhooks for VCS events
package repohooks

import (
	"log/slog"
	"path"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

// defaultEvents are the VCS events that repohooks subscribe to.
var defaultEvents = []vcs.EventType{
	vcs.EventTypePush,
	vcs.EventTypePull,
}

type (
	// hook is a webhook for a VCS repo
	hook struct {
		id            uuid.UUID // internal otf ID
		cloudID       *string   // cloud's hook ID; populated following synchronisation
		vcsProviderID resource.TfeID

		secret    string     // secret token
		repoPath  vcs.Repo   // repo identifier: <repo_owner>/<repo_name>
		vcsKindID vcs.KindID // origin of events
		endpoint  string     // OTF URL that receives events
	}

	newRepohookOptions struct {
		id            *uuid.UUID
		vcsProviderID resource.TfeID
		secret        *string
		repoPath      vcs.Repo
		vcsKindID     vcs.KindID
		cloudID       *string // cloud's webhook id

		// for building endpoint URL
		*internal.HostnameService
	}
)

func newRepohook(opts newRepohookOptions) (*hook, error) {
	hook := hook{
		repoPath:      opts.repoPath,
		vcsKindID:     opts.vcsKindID,
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
	hook.endpoint = opts.WebhookURL(path.Join(handlerPrefix, hook.id.String()))
	return &hook, nil
}

func (h *hook) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", h.id.String()),
		slog.String("vcs_provider_id", h.vcsProviderID.String()),
		slog.String("vcs_kind", string(h.vcsKindID)),
		slog.Any("repo", h.repoPath),
		slog.String("endpoint", h.endpoint),
	}
	if h.cloudID != nil {
		attrs = append(attrs, slog.String("vcs_id", *h.cloudID))
	}
	return slog.GroupValue(attrs...)
}
