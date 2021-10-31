package otf

type StateVersionOutput struct {
	ID string `db:"state_version_output_id"`

	Timestamps

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to State Version
	StateVersionID string
}

type StateVersionOutputList []*StateVersionOutput

func (svo *StateVersionOutput) String() string { return svo.ID }
