package dynamiccreds

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

const gcp provider = "gcp"

var workloadProviderNameRegex = regexp.MustCompile(`^projects/[^/]+/locations/global/workloadIdentityPools/[^/]+/providers/[^/]+$`)

type (
	gcpCredentialConfig struct {
		UniverseDomain   string                    `json:"universe_domain"`
		Type             string                    `json:"type"`
		Audience         string                    `json:"audience"`
		SubjectTokenType string                    `json:"subject_token_type"`
		TokenURL         string                    `json:"token_url"`
		ImpersonationURL string                    `json:"service_account_impersonation_url"`
		CredentialSource gcpCredentialConfigSource `json:"credential_source"`
	}

	gcpCredentialConfigSource struct {
		File   string                          `json:"file"`
		Format gcpCredentialConfigSourceFormat `json:"format"`
	}

	gcpCredentialConfigSourceFormat struct {
		Type string `json:"type"`
	}

	gcpVariablesCredentialsPath struct {
		Credentials string `json:"credentials"`
	}
)

func configureGCP(ctx context.Context, h helper, audience string) (gcpVariablesCredentialsPath, []string, error) {
	serviceAccountEmail, err := h.getRunVar("service_account_email")
	if err != nil {
		return gcpVariablesCredentialsPath{}, nil, err
	}
	// First try "unified variables".
	workloadProviderName, err := h.getVar("workload_provider_name")
	if err != nil {
		// Not found; try separate vars instead.
		projectNumber, err := h.getVar("project_number")
		if err != nil {
			return gcpVariablesCredentialsPath{}, nil, err
		}
		poolID, err := h.getVar("workload_pool_id")
		if err != nil {
			return gcpVariablesCredentialsPath{}, nil, err
		}
		providerID, err := h.getVar("workload_provider_id")
		if err != nil {
			return gcpVariablesCredentialsPath{}, nil, err
		}
		workloadProviderName = fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s/providers/%s",
			projectNumber,
			poolID,
			providerID,
		)
	}
	if !workloadProviderNameRegex.MatchString(workloadProviderName) {
		return gcpVariablesCredentialsPath{}, nil, fmt.Errorf("workload provider name must match the regex: %s", workloadProviderNameRegex.String())
	}
	if audience == "" {
		audience = fmt.Sprintf("//iam.googleapis.com/%s", workloadProviderName)
	}
	token, err := h.generateToken(ctx, audience)
	if err != nil {
		return gcpVariablesCredentialsPath{}, nil, err
	}
	// Construct and write credentials to disk.
	credsPath, err := h.writeFile("token", token)
	if err != nil {
		return gcpVariablesCredentialsPath{}, nil, err
	}
	// Construct and write credentials config.
	credsConfig := gcpCredentialConfig{
		UniverseDomain:   "googleapis.com",
		Type:             "external_account",
		Audience:         audience,
		SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
		TokenURL:         "https://sts.googleapis.com/v1/token",
		ImpersonationURL: fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", serviceAccountEmail),
		CredentialSource: gcpCredentialConfigSource{
			File: credsPath,
			Format: gcpCredentialConfigSourceFormat{
				Type: "text",
			},
		},
	}
	marshaled, err := json.Marshal(credsConfig)
	if err != nil {
		return gcpVariablesCredentialsPath{}, nil, err
	}
	credsConfigPath, err := h.writeFile("config", marshaled)
	if err != nil {
		return gcpVariablesCredentialsPath{}, nil, err
	}
	cfg := gcpVariablesCredentialsPath{
		Credentials: credsConfigPath,
	}
	envs := []string{
		fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", credsConfigPath),
	}
	return cfg, envs, nil
}
