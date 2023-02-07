package team

import (
	"testing"

	"github.com/google/uuid"
)

func NewTestTeam(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return newTeam(uuid.NewString(), org, opts...)
}

func NewTestOwners(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return newTeam("owners", org, opts...)
}
