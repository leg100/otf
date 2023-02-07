package team

import (
	"testing"

	"github.com/google/uuid"
)

func NewTestTeam(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return NewTeam(uuid.NewString(), org, opts...)
}

func NewTestOwners(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return NewTeam("owners", org, opts...)
}
