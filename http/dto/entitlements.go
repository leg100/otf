package dto

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.  Entitlements represents the entitlements of an organization.
type Entitlements struct {
	ID                    string `jsonapi:"primary,entitlement-sets"`
	Agents                bool   `jsonapi:"attr,agents"`
	AuditLogging          bool   `jsonapi:"attr,audit-logging"`
	CostEstimation        bool   `jsonapi:"attr,cost-estimation"`
	Operations            bool   `jsonapi:"attr,operations"`
	PrivateModuleRegistry bool   `jsonapi:"attr,private-module-registry"`
	SSO                   bool   `jsonapi:"attr,sso"`
	Sentinel              bool   `jsonapi:"attr,sentinel"`
	StateStorage          bool   `jsonapi:"attr,state-storage"`
	Teams                 bool   `jsonapi:"attr,teams"`
	VCSIntegrations       bool   `jsonapi:"attr,vcs-integrations"`
}
