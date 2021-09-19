package sqlite

import (
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
	"gorm.io/gorm"
)

// Organization models a row in a organizations table.
type Organization struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	Name                   string
	CollaboratorAuthPolicy tfe.AuthPolicyType
	CostEstimationEnabled  bool
	Email                  string
	OwnersTeamSAMLRoleID   string
	Permissions            *tfe.OrganizationPermissions `gorm:"embedded;embeddedPrefix:permission_"`
	SAMLEnabled            bool
	SessionRemember        int
	SessionTimeout         int
	TrialExpiresAt         time.Time
	TwoFactorConformant    bool
}

// OrganizationList is a list of run models
type OrganizationList []Organization

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *Organization) Update(fn func(*otf.Organization) error) error {
	// model -> domain
	domain := model.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	model.FromDomain(domain)

	return nil
}

func (model *Organization) ToDomain() *otf.Organization {
	domain := otf.Organization{
		ID:                     model.ExternalID,
		Model:                  model.Model,
		Name:                   model.Name,
		CollaboratorAuthPolicy: model.CollaboratorAuthPolicy,
		CostEstimationEnabled:  model.CostEstimationEnabled,
		Email:                  model.Email,
		OwnersTeamSAMLRoleID:   model.OwnersTeamSAMLRoleID,
		Permissions:            model.Permissions,
		SAMLEnabled:            model.SAMLEnabled,
		SessionRemember:        model.SessionRemember,
		SessionTimeout:         model.SessionTimeout,
		TrialExpiresAt:         model.TrialExpiresAt,
		TwoFactorConformant:    model.TwoFactorConformant,
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (model *Organization) FromDomain(org *otf.Organization) {
	model.ExternalID = org.ID
	model.Model = org.Model
	model.Name = org.Name
	model.CollaboratorAuthPolicy = org.CollaboratorAuthPolicy
	model.CostEstimationEnabled = org.CostEstimationEnabled
	model.Email = org.Email
	model.OwnersTeamSAMLRoleID = org.OwnersTeamSAMLRoleID
	model.Permissions = org.Permissions
	model.SAMLEnabled = org.SAMLEnabled
	model.SessionRemember = org.SessionRemember
	model.SessionTimeout = org.SessionTimeout
	model.TrialExpiresAt = org.TrialExpiresAt
	model.TwoFactorConformant = org.TwoFactorConformant
}

func (l OrganizationList) ToDomain() (dl []*otf.Organization) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
