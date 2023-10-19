package github

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/leg100/otf/internal/vcs"
)

func HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *vcs.EventPayload {
	event, err := handleEventWithError(r, secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusAccepted)
	return event
}

func handleEventWithError(r *http.Request, secret string) (*vcs.EventPayload, error) {
	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		return nil, err
	}

	raw, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, err
	}

	// convert github event to an OTF event
	to := vcs.EventPayload{
		VCSKind: vcs.GithubKind,
	}

	switch event := raw.(type) {
	case *github.PushEvent:
		// populate event with list of changed file paths
		for _, c := range event.Commits {
			to.Paths = c.Added
			to.Paths = append(to.Paths, c.Modified...)
			to.Paths = append(to.Paths, c.Removed...)
		}
		to.RepoPath = event.GetRepo().GetFullName()
		to.CommitSHA = event.GetAfter()
		to.CommitURL = event.GetHeadCommit().GetURL()
		to.DefaultBranch = event.GetRepo().GetDefaultBranch()

		to.SenderUsername = event.GetSender().GetLogin()
		to.SenderAvatarURL = event.GetSender().GetAvatarURL()
		to.SenderHTMLURL = event.GetSender().GetHTMLURL()

		if install := event.GetInstallation(); install != nil {
			to.GithubAppInstallID = install.ID
		}

		// a github.PushEvent includes tag events but OTF categorises them as separate
		// event types
		parts := strings.Split(event.GetRef(), "/")
		if len(parts) != 3 || parts[0] != "refs" {
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
		switch parts[1] {
		case "tags":
			to.Type = vcs.EventTypeTag

			switch {
			case event.GetCreated():
				to.Action = vcs.ActionCreated
			case event.GetDeleted():
				to.Action = vcs.ActionDeleted
			default:
				return nil, fmt.Errorf("no action specified for tag event")
			}

			to.Tag = parts[2]

		case "heads":
			to.Type = vcs.EventTypePush
			to.Action = vcs.ActionCreated
			to.Branch = parts[2]

		default:
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
	case *github.PullRequestEvent:
		to.Type = vcs.EventTypePull
		to.RepoPath = event.GetRepo().GetFullName()
		to.PullRequestNumber = event.GetPullRequest().GetNumber()
		to.PullRequestURL = event.GetPullRequest().GetHTMLURL()
		to.PullRequestTitle = event.GetPullRequest().GetTitle()

		to.SenderUsername = event.GetSender().GetLogin()
		to.SenderAvatarURL = event.GetSender().GetAvatarURL()
		to.SenderHTMLURL = event.GetSender().GetHTMLURL()

		if install := event.GetInstallation(); install != nil {
			to.GithubAppInstallID = install.ID
		}

		switch event.GetAction() {
		case "opened":
			to.Action = vcs.ActionCreated
		case "closed":
			if event.PullRequest.GetMerged() {
				to.Action = vcs.ActionMerged
			} else {
				to.Action = vcs.ActionDeleted
			}
		case "synchronize":
			to.Action = vcs.ActionUpdated
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
	case *github.InstallationEvent:
		// ignore events other than uninstallation events
		if event.GetAction() != "deleted" {
			return nil, nil
		}
		to.Action = vcs.ActionDeleted
		to.Type = vcs.EventTypeInstallation
		to.GithubAppInstallID = event.GetInstallation().ID
	default:
		return nil, nil
	}
	if err := to.Validate(); err != nil {
		return nil, err
	}
	return &to, nil
}
