package github

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/pkg/errors"
)

// HandleEvent handles incoming events from github
func HandleEvent(w http.ResponseWriter, r *http.Request, opts otf.HandleEventOptions) otf.VCSEvent {
	event, err := handle(r, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusAccepted)
	return event
}

func handle(r *http.Request, opts otf.HandleEventOptions) (otf.VCSEvent, error) {
	payload, err := github.ValidatePayload(r, []byte(opts.Secret))
	if err != nil {
		return nil, errors.Wrapf(err, "secret: %s", opts.Secret)
	}

	rawEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, err
	}

	switch event := rawEvent.(type) {
	case *github.PushEvent:
		// a github.PushEvent includes tag events but otf categorises them as separate
		// event types
		parts := strings.Split(event.GetRef(), "/")
		if len(parts) != 3 || parts[0] != "refs" {
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
		switch parts[1] {
		case "tags":
			var action otf.VCSTagEventAction
			switch {
			case event.GetCreated():
				action = otf.VCSTagEventCreatedAction
			case event.GetDeleted():
				action = otf.VCSTagEventDeletedAction
			default:
				return nil, fmt.Errorf("no action specified for tag event")
			}

			return &otf.VCSTagEvent{
				WebhookID:  opts.WebhookID,
				Tag:        parts[2],
				Action:     action,
				Identifier: event.GetRepo().GetFullName(),
				CommitSHA:  event.GetAfter(),
			}, nil
		case "heads":
			return &otf.VCSPushEvent{
				WebhookID:  opts.WebhookID,
				Branch:     parts[2],
				Identifier: event.GetRepo().GetFullName(),
				CommitSHA:  event.GetAfter(),
			}, nil
		default:
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
	case *github.PullRequestEvent:
		var action otf.VCSPullEventAction
		switch event.GetAction() {
		case "opened":
			action = otf.VCSPullEventOpened
		case "closed":
			if event.PullRequest.GetMerged() {
				action = otf.VCSPullEventMerged
			} else {
				action = otf.VCSPullEventClosed
			}
		case "synchronised":
			action = otf.VCSPullEventUpdated
		}

		return &otf.VCSPullEvent{
			WebhookID:  opts.WebhookID,
			Action:     action,
			Identifier: event.GetRepo().GetFullName(),
			Branch:     event.PullRequest.Head.GetRef(),
			CommitSHA:  event.GetPullRequest().GetHead().GetSHA(),
		}, nil
	default:
		return nil, nil
	}
}
