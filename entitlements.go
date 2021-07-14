package ots

import tfe "github.com/leg100/go-tfe"

// Entitlements represents the entitlements of an organization.
type Entitlements struct {
	*tfe.Entitlements
}

func (e *Entitlements) DTO() interface{} {
	return e.Entitlements
}

// We currently only support State Storage...
func DefaultEntitlements(organizationID string) *Entitlements {
	return &Entitlements{
		Entitlements: &tfe.Entitlements{
			ID:           organizationID,
			StateStorage: true,
		},
	}
}
