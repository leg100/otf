package otf

import tfe "github.com/leg100/go-tfe"

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.
type Entitlements struct {
	*tfe.Entitlements
}

// DefaultEntitlements constructs an Entitlements struct with currently
// supported entitlements.
func DefaultEntitlements(organizationID string) *Entitlements {
	return &Entitlements{
		Entitlements: &tfe.Entitlements{
			ID:           organizationID,
			StateStorage: true,
			Operations:   true,
		},
	}
}
