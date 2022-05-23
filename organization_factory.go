package otf

import "github.com/google/uuid"

func NewTestOrganization() *Organization {
	ws := Organization{
		id:   NewID("ws"),
		name: uuid.NewString(),
	}
	return &ws
}
