package otf

import "github.com/google/uuid"

func NewTestOrganization() *Organization {
	ws := Organization{
		id:        NewID("ws"),
		createdAt: CurrentTimestamp(),
		updatedAt: CurrentTimestamp(),
		name:      uuid.NewString(),
	}
	return &ws
}
