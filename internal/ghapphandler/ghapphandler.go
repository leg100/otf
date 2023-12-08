// Package ghapphandler provides a handler for the github app webhook endpoint.
package ghapphandler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type Handler struct {
	logr.Logger
	vcs.Publisher

	VCSProviders *vcsprovider.Service
	GithubApps   *github.Service
}

func (h *Handler) AddHandlers(r *mux.Router) {
	r.Handle(github.AppEventsPath, h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// permit handler to talk to services
	ctx := internal.AddSubjectToContext(r.Context(), &internal.Superuser{Username: "github-app-event-handler"})
	// retrieve github app config; if one hasn't been configured then return a
	// 400
	app, err := h.GithubApps.GetApp(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// use github-specific handler to unmarshal event
	payload := github.HandleEvent(w, r, app.WebhookSecret)
	if payload == nil {
		return
	}
	h.V(2).Info("received vcs event", "type", "github-app", "repo", payload.RepoPath)
	// relay a copy of the event for each vcs provider configured with the
	// github app install that triggered the event.
	providers, err := h.VCSProviders.ListVCSProvidersByGithubAppInstall(ctx, *payload.GithubAppInstallID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, prov := range providers {
		h.Publish(vcs.Event{
			EventHeader:  vcs.EventHeader{VCSProviderID: prov.ID},
			EventPayload: *payload,
		})
	}
}
