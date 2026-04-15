package trigger

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

var errTriggerLoop = errors.New("workspace cannot trigger itself")

type trigger struct {
	ID                    resource.TfeID `db:"run_trigger_id"`
	CreatedAt             time.Time      `db:"created_at"`
	WorkspaceID           resource.TfeID `db:"workspace_id"`
	SourceableWorkspaceID resource.TfeID `db:"sourceable_workspace_id"`
}

func newTrigger(workspaceID, sourceableWorkspaceID resource.TfeID) (*trigger, error) {
	if workspaceID == sourceableWorkspaceID {
		return nil, fmt.Errorf("%w: %s", errTriggerLoop, workspaceID)
	}
	return &trigger{
		ID:                    resource.NewTfeID(resource.RunTriggerKind),
		CreatedAt:             internal.CurrentTimestamp(nil),
		WorkspaceID:           workspaceID,
		SourceableWorkspaceID: sourceableWorkspaceID,
	}, nil
}
