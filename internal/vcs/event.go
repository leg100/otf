package vcs

import (
	"errors"
	"fmt"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
)

type EventType string

const (
	EventTypePull         EventType = "pull"
	EventTypePush         EventType = "push"
	EventTypeTag          EventType = "tag"
	EventTypeInstallation EventType = "install" // github-app installation
)

type Action string

const (
	ActionCreated Action = "created"
	ActionDeleted Action = "deleted"
	ActionMerged  Action = "merged"
	ActionUpdated Action = "updated"
)

// ErrIgnoreEvent informs an upstream vcs provider why an event it sent is
// ignored.
type ErrIgnoreEvent struct {
	Reason string
}

func NewErrIgnoreEvent(msg string, args ...any) ErrIgnoreEvent {
	return ErrIgnoreEvent{Reason: fmt.Sprintf(msg, args...)}
}

func (e ErrIgnoreEvent) Error() string {
	return e.Reason
}

type (
	// Event is a VCS event received from a cloud, e.g. a commit event from
	// github
	Event struct {
		EventHeader
		EventPayload
	}

	EventHeader struct {
		// ID of vcs provider that generated this event.
		// event.
		VCSProviderID resource.TfeID
		// Source of vcs provider kind that generated this event.
		Source configversion.Source
	}

	EventPayload struct {
		RepoPath      string
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
)

func (e EventPayload) Validate() error {
	if e.Type == "" {
		return errors.New("event missing event type")
	}
	if e.Action == "" {
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
