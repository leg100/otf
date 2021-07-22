package ots

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StateVersionOutput struct {
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

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
