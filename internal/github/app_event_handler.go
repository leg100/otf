package github

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/vcs"
)

const appEventsPath = "/github-app/events"

// appEventHandler handles events from a github app installation.
type appEventHandler struct {
	Service
	vcs.Publisher
}

func (h *appEventHandler) addHandlers(r *mux.Router) {
	r.HandleFunc(appEventsPath, h.handler)
}

func (h *appEventHandler) handler(w http.ResponseWriter, r *http.Request) {
	app, err := h.GetGithubApp(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event, err := handleEventWithError(r, app.WebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// add installID to event header
	h.Publish(*event)
}
