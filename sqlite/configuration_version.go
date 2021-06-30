package sqlite

import (
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.ConfigurationVersionService = (*ConfigurationVersionService)(nil)

type ConfigurationVersionModel struct {
	gorm.Model

	ExternalID    string
	AutoQueueRuns bool
	Error         string
	ErrorMessage  string
	Source        string
	Speculative   bool
	Status        string
	UploadURL     string

	Configuration []byte
}

type ConfigurationVersionService struct {
	*gorm.DB
}

func NewConfigurationVersionService(db *gorm.DB) *ConfigurationVersionService {
	db.AutoMigrate(&ConfigurationVersionModel{})

	return &ConfigurationVersionService{
		DB: db,
	}
}

func NewConfigurationVersionFromModel(model *ConfigurationVersionModel) *tfe.ConfigurationVersion {
	return &tfe.ConfigurationVersion{
		ID:            model.ExternalID,
		AutoQueueRuns: model.AutoQueueRuns,
		Error:         model.Error,
		ErrorMessage:  model.ErrorMessage,
		Source:        tfe.ConfigurationSource(model.Source),
		Speculative:   model.Speculative,
		Status:        tfe.ConfigurationStatus(model.Status),
		UploadURL:     fmt.Sprintf("/configuration-versions/%s/upload", model.ExternalID),
	}
}

func (ConfigurationVersionModel) TableName() string {
	return "configuration_versions"
}

func (s ConfigurationVersionService) CreateConfigurationVersion(opts *tfe.ConfigurationVersionCreateOptions) (*tfe.ConfigurationVersion, error) {
	model := ConfigurationVersionModel{
		ExternalID:    ots.NewConfigurationVersionID(),
		AutoQueueRuns: ots.DefaultAutoQueueRuns,
		Status:        string(tfe.ConfigurationPending),
	}

	if opts.AutoQueueRuns != nil {
		model.AutoQueueRuns = *opts.AutoQueueRuns
	}

	if opts.Speculative != nil {
		model.Speculative = *opts.Speculative
	}

	if result := s.DB.Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewConfigurationVersionFromModel(&model), nil
}

func (s ConfigurationVersionService) ListConfigurationVersions(opts tfe.ConfigurationVersionListOptions) (*tfe.ConfigurationVersionList, error) {
	var models []ConfigurationVersionModel
	var count int64

	if result := s.DB.Table("configuration_versions").Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := s.DB.Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	var items []*tfe.ConfigurationVersion
	for _, m := range models {
		items = append(items, NewConfigurationVersionFromModel(&m))
	}

	return &tfe.ConfigurationVersionList{
		Items:      items,
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s ConfigurationVersionService) GetConfigurationVersion(id string) (*tfe.ConfigurationVersion, error) {
	model, err := getConfigurationVersionByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewConfigurationVersionFromModel(model), nil
}

func (s ConfigurationVersionService) UploadConfigurationVersion(id string, configuration []byte) error {
	var model ConfigurationVersionModel

	if result := s.DB.Where("external_id = ?", id).First(&model); result.Error != nil {
		return result.Error
	}

	update := map[string]interface{}{
		"configuration": configuration,
		"status":        string(tfe.ConfigurationUploaded),
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return result.Error
	}

	return nil
}

func getConfigurationVersionByID(db *gorm.DB, id string) (*ConfigurationVersionModel, error) {
	var model ConfigurationVersionModel

	if result := db.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
