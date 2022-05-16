package otf

import (
	"encoding/json"
	"time"
)

type StateVersionDBRow struct {
	StateVersionID      string                    `json:"state_version_id"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
	Serial              int32                     `json:"serial"`
	VcsCommitSha        string                    `json:"vcs_commit_sha"`
	VcsCommitUrl        string                    `json:"vcs_commit_url"`
	State               []byte                    `json:"state"`
	RunID               string                    `json:"run_id"`
	StateVersionOutputs []StateVersionOutputDBRow `json:"state_version_outputs"`
}

func UnmarshalStateVersionListFromDB(pgresult interface{}) (stateVersions []*StateVersion, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []StateVersionDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		sv, err := unmarshalStateVersionDBRow(row)
		if err != nil {
			return nil, err
		}
		stateVersions = append(stateVersions, sv)
	}

	return stateVersions, nil
}

func UnmarshalStateVersionFromDB(pgresult interface{}) (*StateVersion, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row StateVersionDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return unmarshalStateVersionDBRow(row)
}

func unmarshalStateVersionDBRow(row StateVersionDBRow) (*StateVersion, error) {
	sv := StateVersion{
		ID: row.StateVersionID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Serial: int64(row.Serial),
		State:  row.State,
		Run:    &Run{ID: row.RunID},
	}

	var err error
	sv.Outputs, err = UnmarshalStateVersionOutputListFromDB(row.StateVersionOutputs)
	if err != nil {
		return nil, err
	}

	return &sv, nil
}
