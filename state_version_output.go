package ots

import (
	"fmt"

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

func NewStateVersionOutputID() string {
	return fmt.Sprintf("wsout-%s", GenerateRandomString(16))
}
