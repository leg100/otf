package forgejo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"

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

	// convert forgejo push event to an OTF event
	to := vcs.EventPayload{VCSKind: vcs.ForgejoKind, Type: vcs.EventTypePull}
	to.RepoPath = event.Repository.FullName
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

	if forgejoClient != nil {

		// get the list of modified file paths
		to.Paths, err = forgejoClient.ListPullRequestFiles(context.TODO(), event.Repository.FullName, int(event.PullRequest.Index))
		if err != nil {
			return nil, err
		}

		// remove duplicate file paths
		slices.Sort(to.Paths)
		to.Paths = slices.Compact(to.Paths)
	} else {
		log.Printf("cannot fetch request files: no client object is available")
	}

	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("failed building OTF event: %w", err)
	}
	return &to, nil
}
