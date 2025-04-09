package organization

import "github.com/leg100/otf/internal/resource"

// OTF is free and therefore the user is entitled to all currently supported
// services.
type Entitlements struct {
	ID                    resource.TfeID
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

// defaultEntitlements constructs an Entitlements struct with currently
// supported entitlements.
func defaultEntitlements(organizationID resource.TfeID) Entitlements {
	return Entitlements{
		ID:                    organizationID,
		Agents:                true,
		AuditLogging:          true,
		CostEstimation:        true,
		Operations:            true,
		PrivateModuleRegistry: true,
		SSO:                   true,
		Sentinel:              true,
		StateStorage:          true,
		Teams:                 true,
		VCSIntegrations:       true,
	}
}
