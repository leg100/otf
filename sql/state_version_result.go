package sql

import (
	"time"

	"github.com/leg100/otf"
)

type stateVersionRow struct {
	StateVersionID      *string               `json:"state_version_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Serial              *int32                `json:"serial"`
	VcsCommitSha        *string               `json:"vcs_commit_sha"`
	VcsCommitUrl        *string               `json:"vcs_commit_url"`
	State               []byte                `json:"state"`
	RunID               *string               `json:"run_id"`
	StateVersionOutputs []StateVersionOutputs `json:"state_version_outputs"`
}

func convertStateVersionList(row FindStateVersionsByWorkspaceNameRow) *otf.StateVersion {
	sv := otf.StateVersion{
		ID: *row.StateVersionID,
		Timestamps: otf.Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Serial: int64(*row.Serial),
		State:  row.State,
		Run:    &otf.Run{ID: *row.RunID},
	}

	for _, svo := range row.StateVersionOutputs {
		sv.Outputs = append(sv.Outputs, convertStateVersionOutput(svo))
	}

	return &sv
}

func convertStateVersion(row stateVersionRow) *otf.StateVersion {
	sv := otf.StateVersion{
		ID: *row.StateVersionID,
		Timestamps: otf.Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Serial: int64(*row.Serial),
		State:  row.State,
		Run:    &otf.Run{ID: *row.RunID},
	}

	for _, svo := range row.StateVersionOutputs {
		sv.Outputs = append(sv.Outputs, convertStateVersionOutput(svo))
	}

	return &sv
}

func convertStateVersionOutput(row StateVersionOutputs) *otf.StateVersionOutput {
	svo := otf.StateVersionOutput{
		ID:        *row.StateVersionOutputID,
		Sensitive: *row.Sensitive,
		Timestamps: otf.Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Type:           *row.Type,
		Value:          *row.Value,
		Name:           *row.Name,
		StateVersionID: *row.StateVersionID,
	}

	return &svo
}
