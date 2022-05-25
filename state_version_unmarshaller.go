package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type StateVersionDBRow struct {
	StateVersionID      string                      `json:"state_version_id"`
	CreatedAt           time.Time                   `json:"created_at"`
	Serial              int                         `json:"serial"`
	VcsCommitSHA        string                      `json:"vcs_commit_sha"`
	VcsCommitURL        string                      `json:"vcs_commit_url"`
	State               []byte                      `json:"state"`
	WorkspaceID         string                      `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

func UnmarshalStateVersionDBResult(row StateVersionDBRow) (*StateVersion, error) {
	sv := StateVersion{
		id:        row.StateVersionID,
		createdAt: row.CreatedAt,
		Serial:    int64(row.Serial),
		State:     row.State,
	}
	for _, r := range row.StateVersionOutputs {
		sv.Outputs = append(sv.Outputs, UnmarshalStateVersionOutputDBType(r))
	}
	return &sv, nil
}
