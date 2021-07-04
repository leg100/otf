package ots

import (
	"errors"
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultRefresh = true
)

var (
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")
)

type RunService interface {
	CreateRun(opts *tfe.RunCreateOptions) (*tfe.Run, error)
	ApplyRun(id string, opts *tfe.RunApplyOptions) error
	GetRun(id string) (*tfe.Run, error)
	ListRuns(workspaceID string, opts tfe.RunListOptions) (*tfe.RunList, error)
	DiscardRun(id string, opts *tfe.RunDiscardOptions) error
	CancelRun(id string, opts *tfe.RunCancelOptions) error
	ForceCancelRun(id string, opts *tfe.RunForceCancelOptions) error
	GetQueuedRuns(opts tfe.RunListOptions) (*tfe.RunList, error)
}

func NewRunID() string {
	return fmt.Sprintf("run-%s", GenerateRandomString(16))
}
