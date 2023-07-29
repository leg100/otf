package cloud

import (
	"github.com/google/uuid"
)

const (
	VCSEventTypePull VCSEventType = iota
	VCSEventTypePush
	VCSEventTypeTag

	VCSActionPullOpened VCSAction = iota
	VCSActionPullClosed
	VCSActionPullMerged
	VCSActionPullUpdated
	VCSActionTagCreated
	VCSActionTagDeleted
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
		Type          VCSEventType
		Action        VCSAction
		Tag           string
		CommitSHA     string
		Branch        string // head branch
		DefaultBranch string

		// Pull request number
		PullNumber int

		// Paths of files that have been added/modified/removed. Only applicable
		// to Push and Tag events types.
		Paths []string
	}

	VCSEventType int
	VCSAction    int
)
