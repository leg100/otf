package trigger

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

var errTriggerLoop = errors.New("workspace cannot trigger itself")

type Trigger struct {
	ID                    resource.TfeID `db:"run_trigger_id"`
	CreatedAt             time.Time      `db:"created_at"`
	WorkspaceID           resource.TfeID `db:"workspace_id"`
	TriggeringWorkspaceID resource.TfeID `db:"triggering_workspace_id"`
}

func newTrigger(workspaceID, triggeringWorkspaceID resource.TfeID) (*Trigger, error) {
	if workspaceID == triggeringWorkspaceID {
		return nil, fmt.Errorf("%w: %s", errTriggerLoop, workspaceID)
	}
	return &Trigger{
		ID:                    resource.NewTfeID(resource.RunTriggerKind),
		CreatedAt:             internal.CurrentTimestamp(nil),
		WorkspaceID:           workspaceID,
		TriggeringWorkspaceID: triggeringWorkspaceID,
	}, nil
}

// LogValue implements slog.LogValuer.
func (t *Trigger) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID.String()),
		slog.Any("workspace_id", t.WorkspaceID),
		slog.Any("triggering_workspace_id", t.TriggeringWorkspaceID),
	)
}
