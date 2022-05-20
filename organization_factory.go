package otf

import "github.com/google/uuid"

func NewTestOrganization() *Organization {
	ws := Organization{
		ID:   NewID("ws"),
		name: uuid.NewString(),
	}
	return &ws
}
