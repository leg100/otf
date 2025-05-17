package workspace

import (
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
)

type Event struct {
	ID           resource.TfeID    `json:"workspace_id"`
	LatestRunID  *resource.TfeID   `json:"latest_run_id"`
	Organization organization.Name `json:"organization_name"`
	LockUsername *user.Username    `json:"lock_username"`
	LockRunID    *resource.TfeID   `json:"lock_run_id"`
}

func (e Event) GetID() resource.TfeID { return e.ID }

func (e *Event) Locked() bool {
	return e.LockUsername != nil || e.LockRunID != nil
}

func (e *Event) Lock() resource.ID {
	if e.LockUsername != nil {
		return e.LockUsername
	}
	if e.LockRunID != nil {
		return e.LockRunID
	}
	return nil
}
