package otf

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
type VCSEvent struct {
	// Repo identifier, <owner>/<repo>
	Identifier      string
	Branch          string
	CommitSHA       string
	IsPullRequest   bool
	OnDefaultBranch bool
	WebhookID       uuid.UUID
}
