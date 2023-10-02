package vcs

type VCSStatus string

const (
	VCSPendingStatus VCSStatus = "pending"
	VCSRunningStatus VCSStatus = "running"
	VCSSuccessStatus VCSStatus = "success"
	VCSErrorStatus   VCSStatus = "error"
	VCSFailureStatus VCSStatus = "failure"
)
