package otf

type StateVersionOutput struct {
	ID string

	Timestamps

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to StateVersion
	StateVersionID string
}

type StateVersionOutputList []*StateVersionOutput

func (svo *StateVersionOutput) String() string { return svo.ID }
