package jsonapi

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.  Entitlements represents the entitlements of an organization.
type Entitlements struct {
	ID                    string `jsonapi:"primary,entitlement-sets"`
	Agents                bool   `jsonapi:"attribute" json:"agents"`
	AuditLogging          bool   `jsonapi:"attribute" json:"audit-logging"`
	CostEstimation        bool   `jsonapi:"attribute" json:"cost-estimation"`
	Operations            bool   `jsonapi:"attribute" json:"operations"`
	PrivateModuleRegistry bool   `jsonapi:"attribute" json:"private-module-registry"`
	SSO                   bool   `jsonapi:"attribute" json:"sso"`
	Sentinel              bool   `jsonapi:"attribute" json:"sentinel"`
	StateStorage          bool   `jsonapi:"attribute" json:"state-storage"`
	Teams                 bool   `jsonapi:"attribute" json:"teams"`
	VCSIntegrations       bool   `jsonapi:"attribute" json:"vcs-integrations"`
}
