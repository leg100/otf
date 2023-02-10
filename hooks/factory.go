package hooks

import (
	"fmt"
	"net/url"
	"path"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

type factory struct {
	cloud.Service
	otf.HostnameService
}

func newFactory(hostnameService otf.HostnameService, cloudService cloud.Service) factory {
	return factory{cloudService, hostnameService}
}

func (f factory) newHook(opts newHookOpts) (*hook, error) {
	cloudConfig, err := f.GetCloudConfig(opts.cloud)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", opts.cloud)
	}

	hook := hook{
		identifier:   opts.identifier,
		cloud:        opts.cloud,
		EventHandler: cloudConfig.Cloud,
		cloudID:      opts.cloudID,
	}

	if opts.id != nil {
		hook.id = *opts.id
	} else {
		hook.id = uuid.New()
	}

	if opts.secret != nil {
		hook.secret = *opts.secret
	} else {
		secret, err := otf.GenerateToken()
		if err != nil {
			return nil, err
		}
		hook.secret = secret
	}

	hook.endpoint = (&url.URL{
		Scheme: "https",
		Host:   f.Hostname(),
		Path:   path.Join(handlerPrefix, hook.id.String()),
	}).String()

	return &hook, nil
}

type newHookOpts struct {
	id         *uuid.UUID
	secret     *string
	identifier string
	cloud      string // cloud name
	cloudID    *string
}
