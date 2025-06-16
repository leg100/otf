package forgejo

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal/vcs"
)

type PushEvent struct {
	Ref        string             `json:"ref"`
	Before     string             `json:"before"`
	After      string             `json:"after"`
	CompareURL string             `json:"compare_url"`
	HeadCommit *forgejo.Commit    `json:"head_commit"`
	Commits    []forgejo.Commit   `json:"commits"`
	Repository forgejo.Repository `json:"repository"`
	Pusher     forgejo.User       `json:"pusher"`
	Sender     forgejo.User       `json:"sender"`
}

func buildUserURL(repourl, reponame, username string) string {
	// workaround: the forgejo.User struct should have an HTMLURL field (it's in the json)
	// foo.com/org/repo → foo.com/user
	return strings.Replace(repourl, reponame, username, 1)
}

func handlePushEvent(b []byte) (*vcs.EventPayload, error) {
	event := PushEvent{}
	err := json.Unmarshal(b, &event)
	if err != nil {
		return nil, err
	}

	// convert forgejo push event to an OTF event
	var to vcs.EventPayload

	repo, err := vcs.NewRepo(event.Repository.Owner.UserName, event.Repository.Name)
	if err != nil {
		return nil, err
	}
	to.Repo = repo

	to.CommitSHA = event.After
	if len(event.Commits) > 0 {
		to.CommitURL = event.Commits[0].URL
	}
	to.DefaultBranch = event.Repository.DefaultBranch
	to.SenderUsername = event.Sender.UserName
	to.SenderAvatarURL = event.Sender.AvatarURL
	to.SenderHTMLURL = buildUserURL(event.Repository.HTMLURL, event.Repository.FullName, event.Sender.UserName)
	// populate event with list of changed file paths
	for _, c := range event.Commits {
		for _, f := range c.Files {
			to.Paths = append(to.Paths, f.Filename)
		}
	}
	// remove duplicate file paths
	slices.Sort(to.Paths)
	to.Paths = slices.Compact(to.Paths)
	// differentiate between tag and branch pushes
	if tag, found := strings.CutPrefix(event.Ref, "refs/tags/"); found {
		to.Type = vcs.EventTypeTag
		to.Tag = tag
		if event.HeadCommit != nil {
			to.Action = vcs.ActionCreated
		} else {
			to.Action = vcs.ActionDeleted
			to.CommitSHA = event.Before
		}
	} else if branch, found := strings.CutPrefix(event.Ref, "refs/heads/"); found {
		to.Type = vcs.EventTypePush
		to.Branch = branch
		// branch pushes are always creates
		to.Action = vcs.ActionCreated
	} else {
		return nil, fmt.Errorf("malformed ref: %s", event.Ref)
	}
	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("failed building OTF event: %w", err)
	}
	return &to, nil
}
