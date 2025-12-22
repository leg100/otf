package github

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/vcs"
)

// AppEventsPath is the URL path for the endpoint receiving VCS events from the
// Github App
const AppEventsPath = "/webhooks/github-app"

type AppEventHandler struct {
	logr.Logger
	vcs.Publisher

	VCSService *vcs.Service
	GithubApps *Service
}

func (h *AppEventHandler) AddHandlers(r *mux.Router) {
	r.Handle(AppEventsPath, h)
}

func (h *AppEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// permit handler to talk to services
	ctx := authz.AddSubjectToContext(r.Context(), &authz.Superuser{Username: "github-app-event-handler"})
	// retrieve github app config; if one hasn't been configured then return a
	// 400
	app, err := h.GithubApps.GetApp(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.V(2).Info("received vcs event", "github_app", app)

	// use github-specific handler to unmarshal event
	payload, err := HandleEvent(r, app.WebhookSecret)
	// either ignore the event, return an error, or publish the event onwards
	var ignore vcs.ErrIgnoreEvent
	if errors.As(err, &ignore) {
		h.V(2).Info("ignoring event: "+err.Error(), "github_app", app)
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		h.Error(err, "handling vcs event", "github_app", app)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// relay a copy of the event for each vcs provider configured with the
	// github app install that triggered the event.
	providers, err := h.VCSService.ListByInstall(ctx, *payload.GithubAppInstallID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, prov := range providers {
		h.Publish(vcs.Event{
			EventHeader: vcs.EventHeader{
				VCSProviderID: prov.ID,
				Source:        Source,
			},
			EventPayload: *payload,
		})
	}
	w.WriteHeader(http.StatusOK)
}
