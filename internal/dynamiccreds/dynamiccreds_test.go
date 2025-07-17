package dynamiccreds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTokenGenerator struct{}

func (f *fakeTokenGenerator) GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	return []byte("fake_oidc_token"), nil
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name     string
		provider provider
		envs     []string
		phase    run.PhaseType
		want     func(t *testing.T, workdir string, gotEnvs []string)
		// wantTfVars returns the expected tfvars file contents. If empty then
		// it is expected that the file does NOT exist.
		wantTfVarsFile func(workdir string) string
	}{
		{
			name:  "no dynamic creds",
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				assert.Empty(t, envs)
			},
		},
		{
			name:     "enable gcp dynamic creds for the default provider",
			provider: gcp,
			envs: []string{
				"TFC_GCP_PROVIDER_AUTH=true",
				"TFC_GCP_WORKLOAD_IDENTITY_AUDIENCE=acme.audience",
				"TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL=terraform@iam.google.com",
				"TFC_GCP_WORKLOAD_PROVIDER_NAME=projects/123/locations/global/workloadIdentityPools/pool-123/providers/provider-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				configPath := filepath.Join(workdir, "gcp_config")
				want := fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", configPath)
				assert.Contains(t, envs, want)
				wantConfig := fmt.Sprintf(`{
  "universe_domain": "googleapis.com",
  "type": "external_account",
  "audience": "acme.audience",
  "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
  "token_url": "https://sts.googleapis.com/v1/token",
  "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/terraform@iam.google.com:generateAccessToken",
  "credential_source": {
    "file": "%s/gcp_token",
    "format": {
      "type": "text"
    }
  }
}`, workdir)
				assertJSONFileMatches(t, wantConfig, configPath)
			},
		},
		{
			name:     "enable gcp dynamic creds for an aliased provider",
			provider: gcp,
			envs: []string{
				"TFC_GCP_PROVIDER_AUTH_ALIAS1=true",
				"TFC_GCP_WORKLOAD_IDENTITY_AUDIENCE_ALIAS1=acme.audience",
				"TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL_ALIAS1=terraform@iam.google.com",
				"TFC_GCP_WORKLOAD_PROVIDER_NAME_ALIAS1=projects/123/locations/global/workloadIdentityPools/pool-123/providers/provider-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				configPath := filepath.Join(workdir, "gcp_ALIAS1_config")
				assert.Empty(t, envs)
				wantConfig := fmt.Sprintf(`{
  "universe_domain": "googleapis.com",
  "type": "external_account",
  "audience": "acme.audience",
  "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
  "token_url": "https://sts.googleapis.com/v1/token",
  "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/terraform@iam.google.com:generateAccessToken",
  "credential_source": {
    "file": "%s/gcp_ALIAS1_token",
    "format": {
      "type": "text"
    }
  }
}`, workdir)
				assertJSONFileMatches(t, wantConfig, configPath)
			},
			wantTfVarsFile: func(workdir string) string {
				return fmt.Sprintf(`{
  "tfc_gcp_dynamic_credentials": {
    "aliases": {
      "ALIAS1": {
        "credentials": "%s/gcp_ALIAS1_config"
      }
    }
  }
}`, workdir)
			},
		},
		{
			name:     "enable aws dynamic creds for the default provider",
			provider: aws,
			envs: []string{
				"TFC_AWS_PROVIDER_AUTH=true",
				"TFC_AWS_RUN_ROLE_ARN=my-arn",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				tokenPath := filepath.Join(workdir, "aws_token")
				assert.FileExists(t, tokenPath)

				assert.Equal(t, []string{
					"AWS_ROLE_ARN=my-arn",
					fmt.Sprintf("AWS_WEB_IDENTITY_TOKEN_FILE=%s", tokenPath),
				}, envs)
			},
		},
		{
			name:     "enable aws dynamic creds for an aliased provider",
			provider: aws,
			envs: []string{
				"TFC_AWS_PROVIDER_AUTH_ALIAS1=true",
				"TFC_AWS_RUN_ROLE_ARN_ALIAS1=my-arn",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				tokenPath := filepath.Join(workdir, "aws_ALIAS1_token")
				assert.FileExists(t, tokenPath)

				assert.Empty(t, envs)

				configPath := filepath.Join(workdir, "aws_ALIAS1_config.ini")
				assert.FileExists(t, configPath)

				wantConfig := fmt.Sprintf(`[default]
web_identity_token_file = %s/aws_ALIAS1_token
role_arn                = my-arn
`, workdir)
				gotConfig := testutils.ReadFile(t, configPath)
				assert.Equal(t, wantConfig, string(gotConfig))

			},
			wantTfVarsFile: func(workdir string) string {
				return fmt.Sprintf(`{
  "tfc_aws_dynamic_credentials": {
    "aliases": {
      "ALIAS1": {
        "shared_config_file": "%s/aws_ALIAS1_config.ini"
      }
    }
  }
}`, workdir)
			},
		},
		{
			name:     "enable azure dynamic creds for the default provider",
			provider: azure,
			envs: []string{
				"TFC_AZURE_PROVIDER_AUTH=true",
				"TFC_AZURE_RUN_CLIENT_ID=clientid-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				assert.Equal(t, []string{
					"ARM_CLIENT_ID=clientid-123",
					"ARM_OIDC_TOKEN=fake_oidc_token",
					"ARM_USE_OIDC=true",
				}, envs)

			},
		},
		{
			name:     "enable azure dynamic creds for an aliased provider",
			provider: azure,
			envs: []string{
				"TFC_AZURE_PROVIDER_AUTH_ALIAS1=true",
				"TFC_AZURE_RUN_CLIENT_ID_ALIAS1=clientid-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				tokenPath := filepath.Join(workdir, "azure_ALIAS1_token")
				assert.FileExists(t, tokenPath)
				assert.Equal(t, "fake_oidc_token", string(testutils.ReadFile(t, tokenPath)))

				clientIDPath := filepath.Join(workdir, "azure_ALIAS1_client_id")
				assert.FileExists(t, clientIDPath)
				assert.Equal(t, "clientid-123", string(testutils.ReadFile(t, clientIDPath)))

				assert.Empty(t, envs)
			},
			wantTfVarsFile: func(workdir string) string {
				return fmt.Sprintf(`{
  "tfc_azure_dynamic_credentials": {
    "aliases": {
      "ALIAS1": {
        "client_id_file_path": "%s/azure_ALIAS1_client_id",
        "oidc_token_file_path": "%s/azure_ALIAS1_token"
      }
    }
  }
}`, workdir, workdir)
			},
		},
		{
			name:     "enable azure dynamic creds for both a default and an aliased provider",
			provider: azure,
			envs: []string{
				"TFC_AZURE_PROVIDER_AUTH=true",
				"TFC_AZURE_RUN_CLIENT_ID=clientid-123",
				"TFC_AZURE_PROVIDER_AUTH_ALIAS1=true",
				"TFC_AZURE_RUN_CLIENT_ID_ALIAS1=clientid-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				tokenPath := filepath.Join(workdir, "azure_token")
				assert.FileExists(t, tokenPath)
				assert.Equal(t, "fake_oidc_token", string(testutils.ReadFile(t, tokenPath)))

				tokenPath = filepath.Join(workdir, "azure_ALIAS1_token")
				assert.FileExists(t, tokenPath)
				assert.Equal(t, "fake_oidc_token", string(testutils.ReadFile(t, tokenPath)))

				clientIDPath := filepath.Join(workdir, "azure_client_id")
				assert.FileExists(t, clientIDPath)
				assert.Equal(t, "clientid-123", string(testutils.ReadFile(t, clientIDPath)))

				clientIDPath = filepath.Join(workdir, "azure_ALIAS1_client_id")
				assert.FileExists(t, clientIDPath)
				assert.Equal(t, "clientid-123", string(testutils.ReadFile(t, clientIDPath)))

				assert.Empty(t, envs)
			},
			wantTfVarsFile: func(workdir string) string {
				return fmt.Sprintf(`{
  "tfc_azure_dynamic_credentials": {
    "default": {
      "client_id_file_path": "%s/azure_client_id",
      "oidc_token_file_path": "%s/azure_token"
    },
    "aliases": {
      "ALIAS1": {
        "client_id_file_path": "%s/azure_ALIAS1_client_id",
        "oidc_token_file_path": "%s/azure_ALIAS1_token"
      }
    }
  }
}`, workdir, workdir, workdir, workdir)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workdir := t.TempDir()
			got, err := Setup(
				context.Background(),
				&fakeTokenGenerator{},
				workdir,
				resource.NewTfeID(resource.JobKind),
				tt.phase,
				tt.envs,
			)
			require.NoError(t, err)
			tt.want(t, workdir, got)
			tfVarsFilename := fmt.Sprintf("%s_dynamic_credentials.auto.tfvars.json", tt.provider)
			tfVarsPath := filepath.Join(workdir, tfVarsFilename)
			if tt.wantTfVarsFile == nil {
				assert.NoFileExists(t, tfVarsPath)
			} else {
				want := tt.wantTfVarsFile(workdir)
				assertJSONFileMatches(t, want, tfVarsPath)
			}
		})
	}
}

func assertJSONFileMatches(t *testing.T, want string, path string) {
	t.Helper()

	assert.FileExists(t, path)
	contents := testutils.ReadFile(t, path)
	var indented bytes.Buffer
	err := json.Indent(&indented, contents, "", "  ")
	require.NoError(t, err)
	assert.Equal(t, want, indented.String())
}
