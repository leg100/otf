package cloud

import (
	"github.com/google/uuid"
)

type VCSEventType int

const (
	VCSPullEventType VCSEventType = iota
	VCSPushEventType
)

// VCSEvent is a VCS event received from a cloud, e.g. a commit event from
// github
type VCSEvent any

// VCSPullEvent occurs when an action is carried out on a pull request
type VCSPullEvent struct {
	RepoID        uuid.UUID
	Action        VCSPullEventAction
	CommitSHA     string
	Branch        string // head branch
	DefaultBranch string
	ChangedPaths  []string
}

type VCSPullEventAction string

const (
	VCSPullEventOpened  VCSPullEventAction = "opened"
	VCSPullEventClosed  VCSPullEventAction = "closed" // closed without merging
	VCSPullEventMerged  VCSPullEventAction = "merged"
	VCSPullEventUpdated VCSPullEventAction = "updated"
)

// VCSPushEvent occurs when a commit is pushed to a repo.
type VCSPushEvent struct {
	RepoID        uuid.UUID
	CommitSHA     string
	Branch        string
	DefaultBranch string
	ChangedPaths  []string
}

// VCSTagEvent occurs when a tag is created or deleted on a repo.
type VCSTagEvent struct {
	RepoID        uuid.UUID
	CommitSHA     string
	Tag           string
	Action        VCSTagEventAction
	DefaultBranch string
}

type VCSTagEventAction string

const (
	VCSTagEventCreatedAction VCSTagEventAction = "created"
	VCSTagEventDeletedAction VCSTagEventAction = "deleted"
)
