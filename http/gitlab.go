package http

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/xanzy/go-gitlab"
)

// GitlabEventHandler handles incoming VCS events from gitlab
type GitlabEventHandler struct {
	token  string
	events chan<- otf.VCSEvent
	logr.Logger
}

func (h *GitlabEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handle(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GitlabEventHandler) handle(r *http.Request) error {
	if token := r.Header.Get("X-Gitlab-Token"); token != h.token {
		return errors.New("token validation failed")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return errors.New("error reading request body")
	}

	event, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
	if err != nil {
		return err
	}

	// Filter out non-push events
	push, ok := event.(*gitlab.PushEvent)
	if !ok {
		return nil
	}

	refParts := strings.Split(push.Ref, "/")
	if len(refParts) != 3 {
		return errors.New("expected ref to be in the format <string>/<string>/<string>")
	}

	h.events <- otf.VCSEvent{
		Identifier: push.Project.PathWithNamespace,
		Branch:     refParts[2],
	}

	return nil
}
