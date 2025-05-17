package run

import (
	"time"

	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/user"
)

type Event struct {
	ID                     resource.TfeID   `json:"run_id"`
	CreatedAt              time.Time        `json:"created_at"`
	CancelSignaledAt       *time.Time       `json:"cancel_signaled_at"`
	IsDestroy              bool             `json:"is_destroy"`
	PositionInQueue        int              `json:"position_in_queue"`
	Refresh                bool             `json:"refresh"`
	RefreshOnly            bool             `json:"refresh_only"`
	ReplaceAddrs           []string         `json:"replace_addrs"`
	TargetAddrs            []string         `json:"target_addrs"`
	Lockfile               []byte           `json:"lock_file"`
	Status                 runstatus.Status `json:"status"`
	WorkspaceID            resource.TfeID   `json:"workspace_id"`
	ConfigurationVersionID resource.TfeID   `json:"configuration_version_id"`
	AutoApply              bool             `json:"auto_apply"`
	PlanOnly               bool             `json:"plan_only"`
	CreatedBy              *user.Username   `json:"created_by"`
	Source                 Source           `json:"source"`
	EngineVersion          string           `json:"engine_version"`
	AllowEmptyApply        bool             `json:"allow_empty_apply"`
	Engine                 *engine.Engine   `json:"engine"`
}

func (e Event) GetID() resource.TfeID { return e.ID }
