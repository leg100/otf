package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
)

type provider string

type TokenGetter interface {
	GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error)
}

// multiConfig permits the configuration of multiple provider blocks: for each
// cloud it permits one for the default provider, and one for each alias
// provider.
// Once marshaled to json and written to the workspace the user can then
// reference these as terraform variables and assign them to each provider
// block.
type multiConfig struct {
	AWS   *providerConfigs[awsVariablesSharedConfigFile]    `json:"tfc_aws_dynamic_credentials"`
	Azure *providerConfigs[azureVariablesCredentialsConfig] `json:"tfc_azure_dynamic_credentials"`
	GCP   *providerConfigs[gcpVariablesCredentialsPath]     `json:"tfc_gcp_dynamic_credentials"`
}

type providerConfigs[T any] struct {
	Default T
	Aliases map[string]T
}

func (c *providerConfigs[T]) addConfig(tag string, config T) {
	if tag == "" {
		c.Default = config
	} else {
		alias := strings.TrimPrefix(tag, "_")
		c.Aliases[alias] = config
	}
}

// Setup cloud providers for dynamic credentials for a job.
func Setup(
	ctx context.Context,
	tokenGetter TokenGetter,
	workdir string,
	jobID resource.TfeID,
	phase run.PhaseType,
	envs map[string]string,
) ([]string, error) {
	var (
		// new environment variables to be returned to the caller.
		newEnvs []string
		cfg     multiConfig
	)

	// Configure dynamic credentials for each cloud provider
	for _, provider := range []provider{aws, azure, gcp} {
		// Enumerate through environment variables
		for k, v := range envs {
			// The user enables dynamic credentials for a cloud provider by
			// specifying an environment variable with a common format. The tag
			// is an empty string if its for the default provider block, or
			// non-empty if it's for an aliased provider block.
			tag, ok := strings.CutPrefix(k, fmt.Sprintf("TFC_%s_PROVIDER_AUTH", strings.ToUpper(string(provider))))
			if !ok {
				continue
			}
			// Must be set to "true" or "false"
			enabled, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			if !enabled {
				continue
			}
			// If a tag is specified then it must be prefixed with an underscore
			// followed by the name of the alias.
			if tag != "" || !strings.HasPrefix(tag, "_") {
				return nil, fmt.Errorf("expected environment variable to have format TFC_<cloud>_PROVIDER_AUTH[_TAG]; instead got: %s", k)
			}

			// Construct helper to assist configuration below.
			h := helper{
				envs:        envs,
				provider:    provider,
				workdir:     workdir,
				jobID:       jobID,
				phase:       phase,
				tag:         tag,
				tokenGetter: tokenGetter,
			}

			// Each cloud provider has environment variables named according to
			// a common pattern from which to read its OIDC audience value.
			audience, err := lookupEnv(envs, provider, tag, "workload_identity_audience")
			if err != nil {
				return nil, err
			}

			// Configure credentials differently depending on the cloud.
			switch provider {
			case aws:
				cloudConfig, envs, err := configureAWS(ctx, h, audience)
				if err != nil {
					return nil, err
				}
				newEnvs = append(newEnvs, envs...)
				cfg.AWS.addConfig(tag, cloudConfig)
			case azure:
				variables, envs, err := configureAzure(ctx, h, audience)
				if err != nil {
					return nil, err
				}
				newEnvs = append(newEnvs, envs...)
				cfg.Azure.addConfig(tag, variables)
			case gcp:
				variables, envs, err := configureGCP(ctx, h, audience)
				if err != nil {
					return nil, err
				}
				newEnvs = append(newEnvs, envs...)
				cfg.GCP.addConfig(tag, variables)
			}
		}
	}
	// write provider config to workspace's workdir.
	{
		marshaled, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(workdir, "dynamic_credentials.auto.tfvars.json")
		if err := os.WriteFile(path, marshaled, 0o644); err != nil {
			return nil, err
		}
	}
	return newEnvs, nil
}

// helper provides helper functions for the configuration of each cloud
// provider.
type helper struct {
	envs        map[string]string
	provider    provider
	workdir     string
	jobID       resource.TfeID
	phase       run.PhaseType
	tag         string
	tokenGetter TokenGetter
}

// getRunVar retrieves from environment variables a run-level value
func (h *helper) getRunVar(name string) (string, error) {
	return lookupRunEnv(h.envs, h.provider, h.tag, h.phase, name)
}

// getVar retrieves from environment variables a value
func (h *helper) getVar(name string) (string, error) {
	return lookupEnv(h.envs, h.provider, h.tag, name)
}

// writeFile writes a file to the workspace's working directory and returns
// its full path.
func (h *helper) writeFile(fname string, data []byte) (string, error) {
	path := filepath.Join(h.workdir, fmt.Sprintf("%s%s_%s", h.provider, h.tag, fname))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// generateToken generates an OIDC token.
func (h *helper) generateToken(ctx context.Context, audience string) ([]byte, error) {
	return h.tokenGetter.GenerateDynamicCredentialsToken(ctx, h.jobID, audience)
}

func lookupEnv(envs map[string]string, provider provider, tag string, name string) (string, error) {
	return tryEnvs(envs,
		fmt.Sprintf("TFC_%s_%s%s", provider, name, tag),
		fmt.Sprintf("TFC_DEFAULT_%s_%s", provider, name),
	)
}

func lookupRunEnv(envs map[string]string, provider provider, tag string, phase run.PhaseType, name string) (string, error) {
	return tryEnvs(envs,
		fmt.Sprintf("TFC_%s_%s_%s%s", provider, strings.ToUpper(string(phase)), name, tag),
		fmt.Sprintf("TFC_DEFAULT_%s_%s_%s", provider, strings.ToUpper(string(phase)), name),
		fmt.Sprintf("TFC_%s_RUN_%s%s", provider, name, tag),
		fmt.Sprintf("TFC_DEFAULT_%s_%s", provider, name),
	)
}

// tryEnvs enumerates names and checks whether there is an environment variable
// with that name and returns the value of the first one found. Otherwise an
// error is returned.
func tryEnvs(envs map[string]string, names ...string) (string, error) {
	for k, v := range envs {
		if slices.Contains(names, k) {
			return v, nil
		}
	}
	return "", fmt.Errorf("at least one of the required environment variables was not found: %v", names)
}
