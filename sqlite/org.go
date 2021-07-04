package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationModel struct {
	gorm.Model

	Name                   string
	ExternalID             string
	Email                  string
	CollaboratorAuthPolicy string
	CostEstimationEnabled  bool
	SessionRemember        int
	SessionTimeout         int
}

type OrganizationService struct {
	*gorm.DB
}

func NewOrganizationService(db *gorm.DB) *OrganizationService {
	db.AutoMigrate(&OrganizationModel{})

	return &OrganizationService{
		DB: db,
	}
}

func NewOrganizationListFromModels(models []OrganizationModel, opts tfe.ListOptions, totalCount int) *tfe.OrganizationList {
	var items []*tfe.Organization
	for _, m := range models {
		items = append(items, NewOrganizationFromModel(&m))
	}

	return &tfe.OrganizationList{
		Items:      items,
		Pagination: ots.NewPagination(opts, totalCount),
	}
}

func NewOrganizationFromModel(model *OrganizationModel) *tfe.Organization {
	return &tfe.Organization{
		Name:                   model.Name,
		ExternalID:             model.ExternalID,
		Email:                  model.Email,
		Permissions:            &ots.DefaultOrganizationPermissions,
		SessionTimeout:         model.SessionTimeout,
		SessionRemember:        model.SessionRemember,
		CollaboratorAuthPolicy: tfe.AuthPolicyType(model.CollaboratorAuthPolicy),
		CostEstimationEnabled:  model.CostEstimationEnabled,
		CreatedAt:              model.CreatedAt,
	}
}

func (OrganizationModel) TableName() string {
	return "organizations"
}

func (s OrganizationService) CreateOrganization(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	org := OrganizationModel{
		Name:                   *opts.Name,
		Email:                  *opts.Email,
		ExternalID:             ots.NewOrganizationID(),
		SessionTimeout:         ots.DefaultSessionTimeout,
		SessionRemember:        ots.DefaultSessionExpiration,
		CollaboratorAuthPolicy: ots.DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  ots.DefaultCostEstimationEnabled,
	}

	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}

	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}

	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = string(*opts.CollaboratorAuthPolicy)
	}

	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
	}

	if result := s.DB.Create(&org); result.Error != nil {
		return nil, result.Error
	}

	return NewOrganizationFromModel(&org), nil
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error) {
	org, err := getOrganizationByName(s.DB, name)
	if err != nil {
		return nil, err
	}

	if opts.Name != nil {
		org.Name = *opts.Name
	}

	if opts.Email != nil {
		org.Email = *opts.Email
	}

	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}

	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}

	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = string(*opts.CollaboratorAuthPolicy)
	}

	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
	}

	if result := s.DB.Save(org); result.Error != nil {
		return nil, result.Error
	}

	return NewOrganizationFromModel(org), nil
}

func (s OrganizationService) ListOrganizations(opts tfe.OrganizationListOptions) (*tfe.OrganizationList, error) {
	var models []OrganizationModel
	var count int64

	if result := s.DB.Table("organizations").Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := s.DB.Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	return NewOrganizationListFromModels(models, opts.ListOptions, int(count)), nil
}

func (s OrganizationService) GetOrganization(name string) (*tfe.Organization, error) {
	model, err := getOrganizationByName(s.DB, name)
	if err != nil {
		return nil, err
	}
	return NewOrganizationFromModel(model), err
}

func (s OrganizationService) DeleteOrganization(name string) error {
	var model OrganizationModel

	if result := s.DB.Where("name = ?", name).First(&model); result.Error != nil {
		return result.Error
	}

	if result := s.DB.Delete(&model); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s OrganizationService) GetEntitlements(name string) (*tfe.Entitlements, error) {
	model, err := getOrganizationByName(s.DB, name)
	if err != nil {
		return nil, err
	}
	return ots.DefaultEntitlements(model.ExternalID), nil
}

func getOrganizationByName(db *gorm.DB, name string) (*OrganizationModel, error) {
	var model OrganizationModel

	if result := db.Where("name = ?", name).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
