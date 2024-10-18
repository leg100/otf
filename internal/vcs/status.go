package vcs

type Status string

const (
	PendingStatus Status = "pending"
	SuccessStatus Status = "success"
	ErrorStatus   Status = "error"
	FailureStatus Status = "failure"
)
