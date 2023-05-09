package notifications

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a notification configuration database on postgres
	pgdb struct {
		internal.DB // provides access to generated SQL queries
	}

	pgresult struct {
		NotificationConfigurationID pgtype.Text        `json:"notification_configuration_id"`
		CreatedAt                   pgtype.Timestamptz `json:"created_at"`
		UpdatedAt                   pgtype.Timestamptz `json:"updated_at"`
		Name                        pgtype.Text        `json:"name"`
		URL                         pgtype.Text        `json:"url"`
		Triggers                    []string           `json:"triggers"`
		DestinationType             pgtype.Text        `json:"destination_type"`
		WorkspaceID                 pgtype.Text        `json:"workspace_id"`
		Enabled                     bool               `json:"enabled"`
	}
)

func (db *pgdb) create(ctx context.Context, ws *Workspace) error {
	_, err := db.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
		ID:                         sql.String(ws.ID),
		CreatedAt:                  sql.Timestamptz(ws.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
		Name:                       sql.String(ws.Name),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		Branch:                     sql.String(ws.Branch),
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		Environment:                sql.String(ws.Environment),
		Description:                sql.String(ws.Description),
		ExecutionMode:              sql.String(string(ws.ExecutionMode)),
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		MigrationEnvironment:       sql.String(ws.MigrationEnvironment),
		SourceName:                 sql.String(ws.SourceName),
		SourceURL:                  sql.String(ws.SourceURL),
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           sql.String(ws.TerraformVersion),
		TriggerPrefixes:            ws.TriggerPrefixes,
		QueueAllRuns:               ws.QueueAllRuns,
		WorkingDirectory:           sql.String(ws.WorkingDirectory),
		OrganizationName:           sql.String(ws.Organization),
	})
	return sql.Error(err)
}
