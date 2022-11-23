package otf

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v41/github"
	"github.com/google/uuid"
)

const GithubEventPathPrefix = "/webhooks/vcs/github"

// GithubEventHandler handles incoming VCS events from github
type GithubEventHandler struct {
	Events chan<- VCSEvent
	logr.Logger

	WebhookSecretGetter
}

func (h *GithubEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawWebhookID := strings.TrimPrefix(r.URL.Path, GithubEventPathPrefix)
	webhookID, err := uuid.Parse(rawWebhookID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.handle(r, webhookID); err != nil {
		h.Error(err, "handling github event", "webhook", webhookID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *GithubEventHandler) handle(r *http.Request, webhookID uuid.UUID) error {
	secret, err := h.GetWebhookSecret(r.Context(), webhookID)
	if err != nil {
		return fmt.Errorf("retrieving webhook secret: %w", err)
	}

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		return fmt.Errorf("validating payload: %w", err)
	}

	rawEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return fmt.Errorf("parsing event: %w", err)
	}

	h.V(2).Info("received github event", "event", github.WebHookType(r))

	switch event := rawEvent.(type) {
	case *github.PushEvent:
		branch, isBranch := ParseBranch(event.GetRef())
		if !isBranch {
			return nil
		}
		h.Events <- VCSEvent{
			Identifier:      event.GetRepo().GetFullName(),
			Branch:          branch,
			CommitSHA:       event.GetAfter(),
			OnDefaultBranch: branch == event.GetRepo().GetDefaultBranch(),
			WebhookID:       webhookID,
		}
	case *github.PullRequestEvent:
		// github pr event ref *is* the branch name, not the standard git ref
		// format refs/branches/<branch>
		branch := event.PullRequest.Head.GetRef()

		h.V(2).Info("forwarding pr event")
		h.Events <- VCSEvent{
			Identifier:    event.GetRepo().GetFullName(),
			Branch:        branch,
			CommitSHA:     event.GetPullRequest().GetHead().GetSHA(),
			IsPullRequest: true,
			WebhookID:     webhookID,
		}
	}

	return nil
}

type WebhookSecretGetter interface {
	GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error)
}
