package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func NewOrganizationFromModel(model *OrganizationModel) *tfe.Organization {
	return &tfe.Organization{
		Name:                   model.Name,
		ExternalID:             model.ExternalID,
		Email:                  model.Email,
		Permissions:            &tfe.OrganizationPermissions{},
		SessionTimeout:         model.SessionTimeout,
		SessionRemember:        model.SessionRemember,
		CollaboratorAuthPolicy: tfe.AuthPolicyType(model.CollaboratorAuthPolicy),
		CostEstimationEnabled:  model.CostEstimationEnabled,
	}
}

func (OrganizationModel) TableName() string {
	return "organizations"
}

func (s OrganizationService) CreateOrganization(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	org, err := ots.NewOrganizationFromOptions(opts)
	if err != nil {
		return nil, err
	}

	model := OrganizationModel{
		Name:                   org.Name,
		ExternalID:             ots.NewOrganizationID(),
		Email:                  org.Email,
		SessionTimeout:         org.SessionTimeout,
		SessionRemember:        org.SessionRemember,
		CollaboratorAuthPolicy: string(org.CollaboratorAuthPolicy),
		CostEstimationEnabled:  org.CostEstimationEnabled,
	}

	if result := s.DB.Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewOrganizationFromModel(&model), nil
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error) {
	var model OrganizationModel

	if result := s.DB.Where("name = ?", name).First(&model); result.Error != nil {
		return nil, result.Error
	}

	update := make(map[string]interface{})
	if opts.Name != nil {
		update["name"] = *opts.Name
	}
	if opts.Email != nil {
		update["email"] = *opts.Email
	}
	if opts.SessionTimeout != nil {
		update["session_timeout"] = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		update["session_remember"] = *opts.SessionRemember
	}
	if opts.CollaboratorAuthPolicy != nil {
		update["collaborator_auth_policy"] = *opts.CollaboratorAuthPolicy
	}
	if opts.CostEstimationEnabled != nil {
		update["cost_estimation_enabled"] = *opts.CostEstimationEnabled
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return nil, result.Error
	}

	return NewOrganizationFromModel(&model), nil
}

func (s OrganizationService) ListOrganizations(opts tfe.OrganizationListOptions) (*tfe.OrganizationList, error) {
	var models []OrganizationModel
	var count int64

	if result := s.DB.Table("organizations").Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := s.DB.Preload(clause.Associations).Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	var items []*tfe.Organization
	for _, m := range models {
		items = append(items, NewOrganizationFromModel(&m))
	}

	return &tfe.OrganizationList{
		Items:      items,
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
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
