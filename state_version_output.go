package otf

type StateVersionOutput struct {
	id string

	Timestamps

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to StateVersion
	StateVersionID string
}

type StateVersionOutputList []*StateVersionOutput

func (svo *StateVersionOutput) ID() string     { return svo.id }
func (svo *StateVersionOutput) String() string { return svo.id }
