package organization

import (
	"time"

	"github.com/leg100/otf/http/jsonapi"
)

var defaultOrganizationPermissions = jsonapiPermissions{
	CanCreateWorkspace: true,
	CanUpdate:          true,
	CanDestroy:         true,
}

type jsonapiOrganization struct {
	Name                  string              `jsonapi:"primary,organizations"`
	CostEstimationEnabled bool                `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt             time.Time           `jsonapi:"attr,created-at,iso8601"`
	ExternalID            string              `jsonapi:"attr,external-id"`
	OwnersTeamSAMLRoleID  string              `jsonapi:"attr,owners-team-saml-role-id"`
	Permissions           *jsonapiPermissions `jsonapi:"attr,permissions"`
	SAMLEnabled           bool                `jsonapi:"attr,saml-enabled"`
	SessionRemember       int                 `jsonapi:"attr,session-remember"`
	SessionTimeout        int                 `jsonapi:"attr,session-timeout"`
	TrialExpiresAt        time.Time           `jsonapi:"attr,trial-expires-at,iso8601"`
	TwoFactorConformant   bool                `jsonapi:"attr,two-factor-conformant"`
}

func (j jsonapiOrganization) toOrganization() *Organization {
	return &Organization{
		id:              j.ExternalID,
		createdAt:       j.CreatedAt,
		name:            j.Name,
		sessionRemember: j.SessionRemember,
		sessionTimeout:  j.SessionTimeout,
	}
}

// jsonapiPermissions represents the organization permissions.
type jsonapiPermissions struct {
	CanCreateTeam               bool `json:"can-create-team"`
	CanCreateWorkspace          bool `json:"can-create-workspace"`
	CanCreateWorkspaceMigration bool `json:"can-create-workspace-migration"`
	CanDestroy                  bool `json:"can-destroy"`
	CanTraverse                 bool `json:"can-traverse"`
	CanUpdate                   bool `json:"can-update"`
	CanUpdateAPIToken           bool `json:"can-update-api-token"`
	CanUpdateOAuth              bool `json:"can-update-oauth"`
	CanUpdateSentinel           bool `json:"can-update-sentinel"`
}

type jsonapiList struct {
	*jsonapi.Pagination
	Items []*jsonapiOrganization
}

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.  Entitlements represents the entitlements of an organization.
type jsonapiEntitlements struct {
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
