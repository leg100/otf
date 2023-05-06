package organization

// OTF is free and therefore the user is entitled to all currently supported
// services.
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

// defaultEntitlements constructs an Entitlements struct with currently
// supported entitlements.
func defaultEntitlements(organizationID string) Entitlements {
	return Entitlements{
		ID:                    organizationID,
		Agents:                true,
		Operations:            true,
		PrivateModuleRegistry: true,
		SSO:                   true,
		StateStorage:          true,
		Teams:                 true,
		VCSIntegrations:       true,
	}
}
