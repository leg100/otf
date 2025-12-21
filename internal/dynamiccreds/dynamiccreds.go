package dynamiccreds

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

type TokenGenerator interface {
	GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error)
}

type providerConfig[T any] struct {
	Default T            `json:"default,omitzero"`
	Aliases map[string]T `json:"aliases,omitzero"`
}

func (c *providerConfig[T]) addConfig(tag string, config T) {
	if tag == "" {
		c.Default = config
	} else {
		if c.Aliases == nil {
			c.Aliases = make(map[string]T)
		}
		alias := strings.TrimPrefix(tag, "_")
		c.Aliases[alias] = config
	}
}

// Setup cloud providers for dynamic credentials for a job.
func Setup(
	ctx context.Context,
	tokenGenerator TokenGenerator,
	workdir string,
	jobID resource.TfeID,
	phase run.PhaseType,
	environmentVariables []string,
) ([]string, error) {
	// new environment variables to be returned to the caller.
	var newEnvs []string

	// convert slice of envs to a map
	envs := make(map[string]string, len(environmentVariables))
	for _, kv := range environmentVariables {
		k, v, found := strings.Cut(kv, "=")
		if !found {
			continue
		}
		envs[k] = v
	}

	// Configure dynamic credentials for each cloud provider
	for _, provider := range []provider{aws, azure, gcp} {
		var (
			providerEnvs []string
			err          error
		)
		switch provider {
		case aws:
			providerEnvs, err = configure(
				ctx,
				envs,
				provider,
				workdir,
				jobID,
				phase,
				tokenGenerator,
				configureAWS,
			)
		case azure:
			providerEnvs, err = configure(
				ctx,
				envs,
				provider,
				workdir,
				jobID,
				phase,
				tokenGenerator,
				configureAzure,
			)
		case gcp:
			providerEnvs, err = configure(
				ctx,
				envs,
				provider,
				workdir,
				jobID,
				phase,
				tokenGenerator,
				configureGCP,
			)
		}
		if err != nil {
			return nil, err
		}
		newEnvs = append(newEnvs, providerEnvs...)
	}
	return newEnvs, nil
}

// configure configures dynamic credentials for a particular cloud provider
func configure[T any](
	ctx context.Context,
	envs map[string]string,
	provider provider,
	workdir string,
	jobID resource.TfeID,
	phase run.PhaseType,
	tokenGenerator TokenGenerator,
	configFunc func(ctx context.Context, h helper, audience string) (T, []string, error),
) ([]string, error) {
	var (
		// environment variables to be added if dynamic credentials are enabled and
		// there are not multiple configs
		newEnvs     []string
		multiConfig providerConfig[T]
		tags        []string
		alias       bool
	)

	// Search environment variables for particular variables that enable
	// dynamic credentials. A collection of "tags" is produced.
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
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		if !b {
			continue
		}
		if tag != "" {
			// If a tag is specified then it must be prefixed with an underscore
			// followed by the name of the alias.
			if !strings.HasPrefix(tag, "_") {
				return nil, fmt.Errorf("expected environment variable to have format TFC_<cloud>_PROVIDER_AUTH[_TAG]; instead got: %s", k)
			}
			alias = true
		}
		tags = append(tags, tag)
	}

	// If at least one alias has been found then a JSON file will be
	// persisted declaring terraform variables for each provider. Otherwise
	// environment variables are added to the environment instead.

	for _, tag := range tags {
		// Construct helper to assist configuration of provider.
		h := helper{
			envs:           envs,
			provider:       provider,
			workdir:        workdir,
			jobID:          jobID,
			phase:          phase,
			tag:            tag,
			TokenGenerator: tokenGenerator,
		}

		// Each cloud provider has environment variables named according to
		// a common pattern from which to read its OIDC audience value. It
		// is optional to provide these variables, so ignore any error.
		audience, _ := lookupEnv(envs, provider, tag, "workload_identity_audience")

		cfg, envs, err := configFunc(ctx, h, audience)
		if err != nil {
			return nil, err
		}

		if alias {
			multiConfig.addConfig(tag, cfg)
		} else {
			newEnvs = append(newEnvs, envs...)
		}
	}
	if alias {
		// write provider config to workspace's workdir.
		variables := map[string]providerConfig[T]{
			fmt.Sprintf("tfc_%s_dynamic_credentials", provider): multiConfig,
		}
		marshaled, err := json.Marshal(variables)
		if err != nil {
			return nil, fmt.Errorf("marshalling variables to json: %w", err)
		}
		fname := fmt.Sprintf("%s_dynamic_credentials.auto.tfvars.json", provider)
		path := filepath.Join(workdir, fname)
		if err := os.WriteFile(path, marshaled, 0o644); err != nil {
			return nil, fmt.Errorf("writing variables file: %w", err)
		}
		return nil, nil
	} else {
		return newEnvs, nil
	}
}

// helper provides helper functions for the configuration of each cloud
// provider.
type helper struct {
	envs     map[string]string
	provider provider
	workdir  string
	jobID    resource.TfeID
	phase    run.PhaseType
	tag      string

	TokenGenerator
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
	return h.GenerateDynamicCredentialsToken(ctx, h.jobID, audience)
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
	for i := range names {
		names[i] = strings.ToUpper(names[i])
	}
	for k, v := range envs {
		if slices.Contains(names, k) {
			return v, nil
		}
	}
	return "", fmt.Errorf("at least one of the required environment variables was not found: %v", names)
}
