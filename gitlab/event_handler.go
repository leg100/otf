package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/xanzy/go-gitlab"
)

// EventHandler handles incoming VCS events from gitlab
type EventHandler struct {
	token  string
	events chan<- otf.VCSEvent
	logr.Logger
}

func (h *EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request, opts otf.HandleEventOptions) otf.VCSEvent {
	if err := h.handle(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *EventHandler) handle(r *http.Request) error {
	if token := r.Header.Get("X-Gitlab-Token"); token != h.token {
		return errors.New("token validation failed")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return errors.New("error reading request body")
	}

	rawEvent, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
	if err != nil {
		return err
	}

	// Filter out non-push events
	switch event := rawEvent.(type) {
	case *gitlab.PushEvent:
		refParts := strings.Split(event.Ref, "/")
		if len(refParts) != 3 {
			return fmt.Errorf("malformed ref: %s", event.Ref)
		}
		h.events <- &otf.VCSPushEvent{}

		//	Identifier: push.Project.PathWithNamespace,
		//	Branch:     refParts[2],
		//}
	case *gitlab.TagEvent:
	case *gitlab.MergeEvent:
	}

	return nil
}
