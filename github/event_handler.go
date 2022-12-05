package github

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
)

// EventHandler handles incoming events from github
type EventHandler struct{}

func (h *EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request, opts otf.HandleEventOptions) *otf.VCSEvent {
	event, err := h.handle(r, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusAccepted)
	return event
}

func (h *EventHandler) handle(r *http.Request, opts otf.HandleEventOptions) (*otf.VCSEvent, error) {
	payload, err := github.ValidatePayload(r, []byte(opts.Secret))
	if err != nil {
		return nil, fmt.Errorf("validating payload: %w", err)
	}

	rawEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}

	switch event := rawEvent.(type) {
	case *github.PushEvent:
		branch, isBranch := otf.ParseBranch(event.GetRef())
		if !isBranch {
			return nil, nil
		}
		return &otf.VCSEvent{
			Identifier:      event.GetRepo().GetFullName(),
			Branch:          branch,
			CommitSHA:       event.GetAfter(),
			OnDefaultBranch: branch == event.GetRepo().GetDefaultBranch(),
			WebhookID:       opts.WebhookID,
		}, nil
	case *github.PullRequestEvent:
		// github pr event ref *is* the branch name, not the standard git ref
		// format refs/branches/<branch>
		branch := event.PullRequest.Head.GetRef()

		return &otf.VCSEvent{
			Identifier:    event.GetRepo().GetFullName(),
			Branch:        branch,
			CommitSHA:     event.GetPullRequest().GetHead().GetSHA(),
			IsPullRequest: true,
			WebhookID:     opts.WebhookID,
		}, nil
	}

	return nil, nil
}
