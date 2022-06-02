package otf

import (
	jsonapi "github.com/leg100/otf/http/dto"
)

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.  Entitlements represents the entitlements of an organization.
type Entitlements struct {
	ID                    string
	Agents                bool
	AuditLogging          bool
	CostEstimation        bool
	Operations            bool
	PrivateModuleRegistry bool
	SSO                   bool
	Sentinel              bool
	StateStorage          bool
	Teams                 bool
	VCSIntegrations       bool
}

// ToJSONAPI assembles a JSONAPI DTO
func (e *Entitlements) ToJSONAPI() any {
	return (*jsonapi.Entitlements)(e)
}

// DefaultEntitlements constructs an Entitlements struct with currently
// supported entitlements.
func DefaultEntitlements(organizationID string) *Entitlements {
	return &Entitlements{
		ID:           organizationID,
		StateStorage: true,
		Operations:   true,
	}
}
