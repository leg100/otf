package vcs

type Status string

const (
	PendingStatus Status = "pending"
	RunningStatus Status = "running"
	SuccessStatus Status = "success"
	ErrorStatus   Status = "error"
	FailureStatus Status = "failure"
)
