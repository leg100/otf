package ots

import (
	"fmt"

	"github.com/google/jsonapi"
)

type StateVersionOutput struct {
	ID        string `jsonapi:"primary,state-version-outputs"`
	Name      string `jsonapi:"attr,name"`
	Sensitive bool   `jsonapi:"attr,sensitive"`
	Type      string `jsonapi:"attr,type"`
	Value     string `jsonapi:"attr,value"`
}

type StateVersionOutputService interface {
	GetStateVersionOutput(id string) (*StateVersionOutput, error)
}

func (svo *StateVersionOutput) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v2/state-version-outputs/%s", svo.ID),
	}
}

func NewStateVersionOutputID() string {
	return fmt.Sprintf("wsout-%s", GenerateRandomString(16))
}
