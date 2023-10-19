package vcs

import "errors"

const (
	EventTypePull EventType = iota + 1
	EventTypePush
	EventTypeTag
	EventTypeInstallation // github-app installation

	ActionCreated Action = iota + 1
	ActionDeleted
	ActionMerged
	ActionUpdated
)

type (
	// Event is a VCS event received from a cloud, e.g. a commit event from
	// github
	Event struct {
		EventHeader
		EventPayload
	}

	EventHeader struct {
		VCSProviderID string
	}

	EventPayload struct {
		RepoPath string

		VCSKind Kind

		Type          EventType
		Action        Action
		Tag           string
		CommitSHA     string
		CommitURL     string
		Branch        string // head branch
		DefaultBranch string

		PullRequestNumber int
		PullRequestURL    string
		PullRequestTitle  string

		SenderUsername  string
		SenderAvatarURL string
		SenderHTMLURL   string

		// Paths of files that have been added/modified/removed. Only applicable
		// to Push and Tag events types.
		Paths []string

		// Only set if event is from a github app
		GithubAppInstallID *int64
	}

	EventType int
	Action    int
)

func (e EventPayload) Validate() error {
	if e.VCSKind == "" {
		return errors.New("event missing vcs kind")
	}
	if e.Type == 0 {
		return errors.New("event missing event type")
	}
	if e.Action == 0 {
		return errors.New("event missing event action")
	}
	switch e.Type {
	case EventTypePush, EventTypePull, EventTypeTag:
		if e.RepoPath == "" {
			return errors.New("event missing repo path")
		}
	}
	return nil
}
