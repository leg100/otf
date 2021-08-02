package sqlite

import (
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
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
func (r *Organization) Update(fn func(*ots.Organization) error) error {
	// model -> domain
	domain := r.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	r.FromDomain(domain)

	return nil
}

func (model *Organization) ToDomain() *ots.Organization {
	domain := ots.Organization{
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
func (r *Organization) FromDomain(org *ots.Organization) {
	r.ExternalID = org.ID
	r.Model = org.Model
	r.Name = org.Name
	r.CollaboratorAuthPolicy = org.CollaboratorAuthPolicy
	r.CostEstimationEnabled = org.CostEstimationEnabled
	r.Email = org.Email
	r.OwnersTeamSAMLRoleID = org.OwnersTeamSAMLRoleID
	r.Permissions = org.Permissions
	r.SAMLEnabled = org.SAMLEnabled
	r.SessionRemember = org.SessionRemember
	r.SessionTimeout = org.SessionTimeout
	r.TrialExpiresAt = org.TrialExpiresAt
	r.TwoFactorConformant = org.TwoFactorConformant
}

func (l OrganizationList) ToDomain() (dl []*ots.Organization) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
