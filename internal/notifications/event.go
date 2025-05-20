package notifications

import "github.com/leg100/otf/internal/resource"

type Event struct {
	ID          resource.TfeID `json:"run_id"`
	WorkspaceID resource.TfeID `json:"workspace_id"`
}
