// Package hooks manages webhooks on VCS repos.
package hooks

import (
	"context"
	"errors"
	"net/url"
	"path"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

var (
	// defaultEvents are the VCS events that hooks subscribe to.
	defaultEvents = []cloud.VCSEventType{
		cloud.VCSPushEventType,
		cloud.VCSPullEventType,
	}
	// errConnected is returned when an attempt is made to delete a
	// webhook but the webhook is still connected e.g. to a module or a
	// workspace.
	errConnected = errors.New("webhook still connected")
)

// synced is a hook that has been synchronised with the cloud
type synced struct {
	*unsynced

	cloud.EventHandler        // handles incoming vcs events
	cloudID            string // cloud's hook ID
}

// unsynced is an hook that is yet to be synchronised with the cloud
type unsynced struct {
	id         uuid.UUID // internal otf ID
	secret     string    // secret token
	identifier string    // repo identifier: <repo_owner>/<repo_name>
	cloud      string    // cloud name
	endpoint   string    // otf URL that receives events
}

// newHook constructs an unsynchronised hook - a hook always begins its life
// in an unsynchronised state.
func newHook(opts newHookOpts) (*unsynced, error) {
	hook := unsynced{
		identifier: opts.identifier,
		cloud:      opts.cloud,
	}

	if opts.id != nil {
		hook.id = *opts.id
	} else {
		hook.id = uuid.New()
	}

	if opts.secret != nil {
		hook.secret = *opts.secret
	} else {
		secret, err := otf.GenerateToken()
		if err != nil {
			return nil, err
		}
		hook.secret = secret
	}

	hook.endpoint = (&url.URL{
		Scheme: "https",
		Host:   opts.hostname,
		Path:   path.Join("/webhooks/vcs", hook.id.String()),
	}).String()

	return &hook, nil
}

type newHookOpts struct {
	id         *uuid.UUID
	secret     *string
	identifier string
	cloud      string // cloud name
	hostname   string
}

// db is a database for hooks
type db interface {
	create(context.Context, *unsynced, syncFunc) (*synced, error)
	get(context.Context, uuid.UUID) (*synced, error)
	delete(context.Context, uuid.UUID) (*synced, error)
}

// syncFunc synchronises config between DB and cloud, taking an existing hook
// (or nil if one doesn't exist yet) and a DB transaction wrapper, and returning
// the cloud provider's hook ID
type syncFunc func(hook *synced, tx otf.Database) (string, error)
