package sql

import (
	"context"

	"github.com/leg100/otf"
)

// GetLogs retrieves the logs for a given run phase.
func (db *DB) GetLogs(ctx context.Context, runID string, phase otf.PhaseType) ([]byte, error) {
	data, err := db.FindLogs(ctx, String(runID), String(string(phase)))
	if err != nil {
		// Don't consider no rows an error because logs may not have been
		// uploaded yet.
		if noRowsInResultError(err) {
			return nil, nil
		}
		return nil, Error(err)
	}
	return data, nil
}
