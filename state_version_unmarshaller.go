package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type StateVersionDBRow struct {
	StateVersionID      string                      `json:"state_version_id"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Serial              int                         `json:"serial"`
	VcsCommitSHA        string                      `json:"vcs_commit_sha"`
	VcsCommitURL        string                      `json:"vcs_commit_url"`
	State               []byte                      `json:"state"`
	WorkspaceID         string                      `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

func UnmarshalStateVersionDBResult(row StateVersionDBRow) (*StateVersion, error) {
	sv := StateVersion{
		ID: row.StateVersionID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Serial: int64(row.Serial),
		State:  row.State,
	}

	for _, r := range row.StateVersionOutputs {
		out, err := UnmarshalStateVersionOutputDBType(r)
		if err != nil {
			return nil, err
		}
		sv.Outputs = append(sv.Outputs, out)
	}

	return &sv, nil
}
