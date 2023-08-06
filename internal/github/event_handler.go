package github

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf/internal/cloud"
)

// HandleEvent handles incoming events from github
func HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *cloud.VCSEvent {
	event, err := handle(r, secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusAccepted)
	return event
}

func handle(r *http.Request, secret string) (*cloud.VCSEvent, error) {
	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		return nil, err
	}

	rawEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, err
	}

	// convert github event to an OTF event
	to := cloud.VCSEvent{
		Cloud: cloud.Github,
	}

	switch event := rawEvent.(type) {
	case *github.PushEvent:
		// populate event with list of changed file paths
		for _, c := range event.Commits {
			to.Paths = c.Added
			to.Paths = append(to.Paths, c.Modified...)
			to.Paths = append(to.Paths, c.Removed...)
		}
		to.CommitSHA = event.GetAfter()
		to.CommitURL = event.GetHeadCommit().GetURL()
		to.DefaultBranch = event.GetRepo().GetDefaultBranch()

		to.SenderUsername = event.GetSender().GetLogin()
		to.SenderAvatarURL = event.GetSender().GetAvatarURL()
		to.SenderHTMLURL = event.GetSender().GetHTMLURL()

		// a github.PushEvent includes tag events but OTF categorises them as separate
		// event types
		parts := strings.Split(event.GetRef(), "/")
		if len(parts) != 3 || parts[0] != "refs" {
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
		switch parts[1] {
		case "tags":
			to.Type = cloud.VCSEventTypeTag

			switch {
			case event.GetCreated():
				to.Action = cloud.VCSActionCreated
			case event.GetDeleted():
				to.Action = cloud.VCSActionDeleted
			default:
				return nil, fmt.Errorf("no action specified for tag event")
			}

			to.Tag = parts[2]

			return &to, nil
		case "heads":
			to.Type = cloud.VCSEventTypePush
			to.Action = cloud.VCSActionCreated
			to.Branch = parts[2]

			return &to, nil
		default:
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
	case *github.PullRequestEvent:
		to.Type = cloud.VCSEventTypePull
		to.PullRequestNumber = event.GetPullRequest().GetNumber()
		to.PullRequestURL = event.GetPullRequest().GetHTMLURL()
		to.PullRequestTitle = event.GetPullRequest().GetTitle()

		to.SenderUsername = event.GetSender().GetLogin()
		to.SenderAvatarURL = event.GetSender().GetAvatarURL()
		to.SenderHTMLURL = event.GetSender().GetHTMLURL()

		switch event.GetAction() {
		case "opened":
			to.Action = cloud.VCSActionCreated
		case "closed":
			if event.PullRequest.GetMerged() {
				to.Action = cloud.VCSActionMerged
			} else {
				to.Action = cloud.VCSActionDeleted
			}
		case "synchronize":
			to.Action = cloud.VCSActionUpdated
		default:
			// ignore other pull request events
			return nil, nil
		}

		to.Branch = event.PullRequest.Head.GetRef()
		to.CommitSHA = event.GetPullRequest().GetHead().GetSHA()
		to.DefaultBranch = event.GetRepo().GetDefaultBranch()

		// commit-url isn't provided in a pull-request event so one is
		// constructed instead
		to.CommitURL = event.GetRepo().GetHTMLURL() + "/commit/" + to.CommitSHA

		return &to, nil
	default:
		return nil, nil
	}
}
