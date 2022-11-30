package otf

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
)

// GithubEventHandler handles incoming events from github
type GithubEventHandler struct{}

func (h *GithubEventHandler) HandleEvent(w http.ResponseWriter, r *http.Request, hook *Webhook) *VCSEvent {
	event, err := h.handle(r, hook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusAccepted)
	return event
}

func (h *GithubEventHandler) handle(r *http.Request, hook *Webhook) (*VCSEvent, error) {
	payload, err := github.ValidatePayload(r, []byte(hook.Secret))
	if err != nil {
		return nil, fmt.Errorf("validating payload: %w", err)
	}

	rawEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}

	switch event := rawEvent.(type) {
	case *github.PushEvent:
		branch, isBranch := ParseBranch(event.GetRef())
		if !isBranch {
			return nil, nil
		}
		return &VCSEvent{
			Identifier:      event.GetRepo().GetFullName(),
			Branch:          branch,
			CommitSHA:       event.GetAfter(),
			OnDefaultBranch: branch == event.GetRepo().GetDefaultBranch(),
			WebhookID:       hook.WebhookID,
		}, nil
	case *github.PullRequestEvent:
		// github pr event ref *is* the branch name, not the standard git ref
		// format refs/branches/<branch>
		branch := event.PullRequest.Head.GetRef()

		return &VCSEvent{
			Identifier:    event.GetRepo().GetFullName(),
			Branch:        branch,
			CommitSHA:     event.GetPullRequest().GetHead().GetSHA(),
			IsPullRequest: true,
			WebhookID:     hook.WebhookID,
		}, nil
	}

	return nil, nil
}
