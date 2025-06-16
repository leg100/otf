package forgejo

import (
	"encoding/json"
	"fmt"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal/vcs"
)

type PullRequestEvent struct {
	Action      string              `json:"action"`
	CommitID    string              `json:"commit_id"`
	Number      int64               `json:"number"`
	PullRequest forgejo.PullRequest `json:"pull_request"`
	Repository  forgejo.Repository  `json:"repository"`
	Sender      forgejo.User        `json:"sender"`
}

func forgejoActionToVcs(pr *PullRequestEvent) vcs.Action {
	rv := map[string]vcs.Action{
		"opened":   vcs.ActionCreated,
		"reopened": vcs.ActionCreated,
		"closed":   vcs.ActionDeleted, // or ActionMerged; checked below
	}[pr.Action]
	if rv == "" {
		// catchall (edited/assigned/unassigned/label_updated/label_cleared/synchronized/milestoned/demilestoned/reviewed/review_requested/review_request_removed)
		rv = vcs.ActionUpdated
	}
	if rv == vcs.ActionDeleted && pr.PullRequest.HasMerged {
		rv = vcs.ActionMerged
	}
	return vcs.Action(rv)
}

func handlePullRequestEvent(b []byte) (*vcs.EventPayload, error) {
	event := PullRequestEvent{}
	err := json.Unmarshal(b, &event)
	if err != nil {
		return nil, err
	}

	// convert forgejo PR event to an OTF event
	to := vcs.EventPayload{Type: vcs.EventTypePull}

	repo, err := vcs.NewRepo(event.Repository.Owner.UserName, event.Repository.Name)
	if err != nil {
		return nil, err
	}
	to.Repo = repo

	to.Branch = event.PullRequest.Head.Name
	to.CommitSHA = event.CommitID
	if to.CommitSHA == "" && event.PullRequest.Head != nil {
		to.CommitSHA = event.PullRequest.Head.Sha
	}
	to.Action = forgejoActionToVcs(&event)
	to.PullRequestNumber = int(event.PullRequest.Index)
	to.PullRequestURL = event.PullRequest.HTMLURL
	to.PullRequestTitle = event.PullRequest.Title

	to.DefaultBranch = event.Repository.DefaultBranch
	to.SenderUsername = event.Sender.UserName
	to.SenderAvatarURL = event.Sender.AvatarURL
	to.SenderHTMLURL = buildUserURL(event.Repository.HTMLURL, event.Repository.FullName, event.Sender.UserName)

	to.Paths = nil // it will be looked up later

	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("failed building OTF event: %w", err)
	}
	return &to, nil
}
