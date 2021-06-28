package ots

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
)

type StateVersionOutputService interface {
	GetStateVersionOutput(id string) (*tfe.StateVersionOutput, error)
}

func NewStateVersionOutputID() string {
	return fmt.Sprintf("wsout-%s", GenerateRandomString(16))
}
