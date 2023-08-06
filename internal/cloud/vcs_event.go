package cloud

import (
	"github.com/google/uuid"
)

const (
	VCSEventTypePull VCSEventType = iota
	VCSEventTypePush
	VCSEventTypeTag

	VCSActionCreated VCSAction = iota
	VCSActionDeleted
	VCSActionMerged
	VCSActionUpdated
)

type (
	// VCSEvent is a VCS event received from a cloud, e.g. a commit event from
	// github
	VCSEvent struct {
		//
		// These fields are populated by the generic webhook handler
		//
		RepoID        uuid.UUID
		VCSProviderID string
		RepoPath      string

		//
		// These fields are populated by cloud-specific handlers
		//
		Cloud Kind

		Type          VCSEventType
		Action        VCSAction
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

	VCSEventType int
	VCSAction    int
)
