package ots

import (
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

type ApplyService interface {
	GetApply(id string) (*tfe.Apply, error)
}

func NewApplyID() string {
	return fmt.Sprintf("apply-%s", GenerateRandomString(16))
}
