package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
)

const (
	// HandlerPrefix is the URL path prefix for the endpoint receiving vcs events
	HandlerPrefix = "/webhooks/vcs"

	WebhookContextKey key = 0
)

type (
	// handler is the first point of entry for incoming VCS events, relaying them onto
	// a cloud-specific handler.
	handler struct {
		logr.Logger

		handlerBroker
		handlerDB
	}

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*Hook, error)
	}

	handlerBroker interface {
		Publish(cloud.VCSEvent)
	}

	// cloudHandler extracts a cloud-specific event from the http request, converting it into a
	// VCS event. Returns nil if the event is to be ignored.
	cloudHandler interface {
		HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *cloud.VCSEvent
	}

	key int
)

func (h *handler) AddHandlers(r *mux.Router) {
	r = r.Path(path.Join(HandlerPrefix, "{webhook_id}")).Subrouter()
	r.Use(h.getHook)
}

func (h *handler) getHook(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opts struct {
			ID uuid.UUID `schema:"webhook_id,required"`
		}
		if err := decode.All(&opts, r); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		hook, err := h.getHookByID(r.Context(), opts.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.V(1).Info("received vcs event", "id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

		if event := hook.HandleEvent(w, r, hook.secret); event != nil {
			// add non-cloud specific info to event before publishing
			event.RepoID = hook.id
			event.RepoPath = hook.identifier
			// TODO: set oauth-token-id instead of vcs provider id
			event.VCSProviderID = hook.vcsProviderID

			h.Publish(*event)
		}
	})
}
