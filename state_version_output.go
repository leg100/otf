package otf

import (
	"gorm.io/gorm"
)

type StateVersionOutput struct {
	ID string

	gorm.Model

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to State Version
	StateVersionID uint
}

type StateVersionOutputList []*StateVersionOutput
