package github

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal/vcs"
)

func HandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		return nil, fmt.Errorf("validating payload: %w", err)
	}
	raw, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("parsing payload: %w", err)
	}

	// convert github event to an OTF event
	var to vcs.EventPayload
	switch event := raw.(type) {
	case *github.PushEvent:
		to.Repo = vcs.Repo{Owner: event.GetRepo().Owner.GetName(), Name: event.GetRepo().GetName()}
		to.CommitSHA = event.GetAfter()
		to.CommitURL = event.GetHeadCommit().GetURL()
		to.DefaultBranch = event.GetRepo().GetDefaultBranch()
		to.SenderUsername = event.GetSender().GetLogin()
		to.SenderAvatarURL = event.GetSender().GetAvatarURL()
		to.SenderHTMLURL = event.GetSender().GetHTMLURL()
		if install := event.GetInstallation(); install != nil {
			to.GithubAppInstallID = install.ID
		}
		// populate event with list of changed file paths
		for _, c := range event.Commits {
			to.Paths = append(to.Paths, c.Added...)
			to.Paths = append(to.Paths, c.Modified...)
			to.Paths = append(to.Paths, c.Removed...)
		}
		// remove duplicate file paths
		slices.Sort(to.Paths)
		to.Paths = slices.Compact(to.Paths)
		// differentiate between tag and branch pushes
		if tag, found := strings.CutPrefix(event.GetRef(), "refs/tags/"); found {
			to.Type = vcs.EventTypeTag
			to.Tag = tag
			// tags can be created or deleted
			switch {
			case event.GetCreated():
				to.Action = vcs.ActionCreated
			case event.GetDeleted():
				to.Action = vcs.ActionDeleted
			default:
				return nil, fmt.Errorf("no action found for tag event")
			}
		} else if branch, found := strings.CutPrefix(event.GetRef(), "refs/heads/"); found {
			to.Type = vcs.EventTypePush
			to.Branch = branch
			// branch pushes are always a create
			to.Action = vcs.ActionCreated
		} else {
			return nil, fmt.Errorf("malformed ref: %s", event.GetRef())
		}
	case *github.PullRequestEvent:
		to.Type = vcs.EventTypePull
		to.Repo = vcs.Repo{Owner: event.GetRepo().Owner.GetLogin(), Name: event.GetRepo().GetName()}
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
			return nil, vcs.NewErrIgnoreEvent("unsupported action: %s", event.GetAction())
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
			return nil, vcs.NewErrIgnoreEvent("unsupported action: %s", event.GetAction())
		}
		to.Action = vcs.ActionDeleted
		to.Type = vcs.EventTypeInstallation
		to.GithubAppInstallID = event.GetInstallation().ID
	default:
		return nil, vcs.NewErrIgnoreEvent("unsupported event: %T", raw)
	}
	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("failed building OTF event: %w", err)
	}
	return &to, nil
}
