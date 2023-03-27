package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/leg100/otf/cloud"
	"github.com/xanzy/go-gitlab"
)

func HandleEvent(w http.ResponseWriter, r *http.Request, opts cloud.HandleEventOptions) cloud.VCSEvent {
	event, err := handle(r, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	return event
}

func handle(r *http.Request, opts cloud.HandleEventOptions) (cloud.VCSEvent, error) {
	if token := r.Header.Get("X-Gitlab-Token"); token != opts.Secret {
		return nil, errors.New("token validation failed")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, errors.New("error reading request body")
	}

	rawEvent, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
	if err != nil {
		return nil, err
	}

	switch event := rawEvent.(type) {
	case *gitlab.PushEvent:
		refParts := strings.Split(event.Ref, "/")
		if len(refParts) != 3 {
			return nil, fmt.Errorf("malformed ref: %s", event.Ref)
		}
		return cloud.VCSPushEvent{
			RepoID:        opts.RepoID,
			Branch:        refParts[2],
			CommitSHA:     event.After,
			DefaultBranch: event.Project.DefaultBranch,
		}, nil
	case *gitlab.TagEvent:
		refParts := strings.Split(event.Ref, "/")
		if len(refParts) != 3 {
			return nil, fmt.Errorf("malformed ref: %s", event.Ref)
		}
		return cloud.VCSTagEvent{
			RepoID: opts.RepoID,
			Tag:    refParts[2],
			// Action:     action,
			CommitSHA:     event.After,
			DefaultBranch: event.Project.DefaultBranch,
		}, nil
	case *gitlab.MergeEvent:
	}

	return nil, nil
}
