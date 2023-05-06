package inmem

import (
	"fmt"

	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
)

type CloudService struct {
	db map[string]cloud.Config // keyed by cloud name
}

func NewCloudService(configs ...cloud.Config) (*CloudService, error) {
	db := make(map[string]cloud.Config, len(configs))
	for _, cfg := range configs {
		db[cfg.Name] = cfg
	}
	return &CloudService{db}, nil
}

func (cs *CloudService) GetCloudConfig(name string) (cloud.Config, error) {
	cfg, ok := cs.db[name]
	if !ok {
		return cloud.Config{}, fmt.Errorf("unknown cloud: %s", cfg)
	}
	return cfg, nil
}

func (cs *CloudService) ListCloudConfigs() []cloud.Config {
	var configs []cloud.Config
	for _, cfg := range cs.db {
		configs = append(configs, cfg)
	}
	return configs
}

func NewCloudServiceWithDefaults() *CloudService {
	return &CloudService{
		db: map[string]cloud.Config{
			"github": github.Defaults(),
			"gitlab": gitlab.Defaults(),
		},
	}
}
