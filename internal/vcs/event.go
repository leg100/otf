package vcs

import (
	"github.com/google/uuid"
	"github.com/leg100/otf/internal/cloud"
)

const (
	EventTypePull EventType = iota
	EventTypePush
	EventTypeTag

	ActionCreated Action = iota
	ActionDeleted
	ActionMerged
	ActionUpdated
)

type (
	// Event is a VCS event received from a cloud, e.g. a commit event from
	// github
	Event struct {
		//
		// These fields are populated by the generic webhook handler
		//
		RepoID        uuid.UUID
		VCSProviderID string
		RepoPath      string

		//
		// These fields are populated by cloud-specific handlers
		//
		Cloud cloud.Kind

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
	}

	EventType int
	Action    int
)
