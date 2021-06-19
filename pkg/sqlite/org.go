package sqlite

import (
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationModel struct {
	gorm.Model

	Name                   string
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

func NewOrganizationFromModel(model *OrganizationModel) *ots.Organization {
	return &ots.Organization{
		Name:                   model.Name,
		Email:                  model.Email,
		Permissions:            &ots.OrganizationPermissions{},
		SessionTimeout:         model.SessionTimeout,
		SessionRemember:        model.SessionRemember,
		CollaboratorAuthPolicy: tfe.AuthPolicyType(model.CollaboratorAuthPolicy),
		CostEstimationEnabled:  model.CostEstimationEnabled,
	}
}

func (OrganizationModel) TableName() string {
	return "organizations"
}

func (s OrganizationService) CreateOrganization(opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
	org, err := ots.NewOrganizationFromOptions(opts)
	if err != nil {
		return nil, err
	}

	model := OrganizationModel{
		Name:                   org.Name,
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

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
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

func (s OrganizationService) ListOrganizations(opts ots.OrganizationListOptions) (*ots.OrganizationList, error) {
	var models []OrganizationModel

	if result := s.DB.Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	orgs := &ots.OrganizationList{
		OrganizationListOptions: ots.OrganizationListOptions{
			ListOptions: opts.ListOptions,
		},
	}
	for _, m := range models {
		orgs.Items = append(orgs.Items, NewOrganizationFromModel(&m))
	}

	return orgs, nil
}

func (s OrganizationService) GetOrganization(name string) (*ots.Organization, error) {
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

func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	return &ots.Entitlements{}, nil
}

func getOrganizationByName(db *gorm.DB, name string) (*OrganizationModel, error) {
	var model OrganizationModel

	if result := db.Where("name = ?", name).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
