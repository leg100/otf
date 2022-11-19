package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v41/github"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// GithubEventHandler handles incoming VCS events from github
type GithubEventHandler struct {
	secret string
	events chan<- otf.VCSEvent
	logr.Logger
}

func (h *GithubEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handle(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *GithubEventHandler) handle(r *http.Request) error {
	payload, err := github.ValidatePayload(r, []byte(h.secret))
	if err != nil {
		h.Error(err, "received invalid github event payload")
		return err
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return err
	}

	// Filter out non-push events
	push, ok := event.(*github.PushEvent)
	if !ok {
		return nil
	}

	refParts := strings.Split(push.GetRef(), "/")
	if len(refParts) != 3 {
		return errors.New("expected ref to be in the format <string>/<string>/<string>")
	}

	h.events <- otf.VCSEvent{
		OrganizationName: mux.Vars(r)["organization_name"],
		WorkspaceName:    mux.Vars(r)["workspace_name"],
		Identifier:       push.GetRepo().GetFullName(),
		Branch:           refParts[2],
	}
	return nil
}
