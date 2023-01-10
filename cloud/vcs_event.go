package cloud

import (
	"github.com/google/uuid"
)

type VCSEventType int

const (
	VCSPullEventType VCSEventType = iota
	VCSPushEventType
)

// VCSEvent is an event received from a VCS provider, e.g. a commit event from
// github
type VCSEvent any

// VCSPullEvent occurs when an action is carried out on a pull request
type VCSPullEvent struct {
	WebhookID  uuid.UUID
	Action     VCSPullEventAction
	Identifier string // repo identifier, <owner>/<repo>
	CommitSHA  string
	Branch     string // head branch
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
	WebhookID  uuid.UUID
	Identifier string // repo identifier, <owner>/<repo>
	CommitSHA  string
	Branch     string
}

// VCSTagEvent occurs when a tag is created or deleted on a repo.
type VCSTagEvent struct {
	WebhookID  uuid.UUID
	Identifier string // repo identifier, <owner>/<repo>
	CommitSHA  string
	Tag        string
	Action     VCSTagEventAction
}

type VCSTagEventAction string

const (
	VCSTagEventCreatedAction VCSTagEventAction = "created"
	VCSTagEventDeletedAction VCSTagEventAction = "deleted"
)
