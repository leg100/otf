package inmem

import (
	"fmt"

	"github.com/leg100/otf"
)

type CloudService struct {
	db map[string]otf.CloudConfig // keyed by cloud name
}

func NewCloudService(configs ...otf.CloudConfig) (*CloudService, error) {
	db := make(map[string]otf.CloudConfig, len(configs))
	for _, cfg := range configs {
		db[cfg.Name] = cfg
	}
	return &CloudService{db}, nil
}

// TODO: rename to GetCloudConfig
func (cs *CloudService) GetCloud(name string) (otf.CloudConfig, error) {
	cfg, ok := cs.db[name]
	if !ok {
		return otf.CloudConfig{}, fmt.Errorf("unknown cloud: %s", cfg)
	}
	return cfg, nil
}

// TODO: rename to ListCloudConfigs()
func (cs *CloudService) ListClouds() []otf.CloudConfig {
	var configs []otf.CloudConfig
	for _, cfg := range cs.db {
		configs = append(configs, cfg)
	}
	return configs
}

func NewTestCloudService() *CloudService {
	return &CloudService{
		db: map[string]otf.CloudConfig{
			"github": otf.GithubDefaults(),
			"gitlab": otf.GitlabDefaults(),
		},
	}
}
