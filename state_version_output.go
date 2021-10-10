package otf

type StateVersionOutput struct {
	ID string

	Model

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to State Version
	StateVersionID int64
}

type StateVersionOutputList []*StateVersionOutput
