package dynamiccreds

import (
	"context"
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTokenGenerator struct{}

func (f *fakeTokenGenerator) GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	return []byte("fake_oidc_token"), nil
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name  string
		envs  []string
		phase run.PhaseType
		want  func(t *testing.T, workdir string, gotEnvs []string)
	}{
		{
			name:  "no dynamic creds",
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				assert.Empty(t, envs)
			},
		},
		{
			name: "enable gcp dynamic creds",
			envs: []string{
				"TFC_GCP_PROVIDER_AUTH=true",
				"TFC_GCP_WORKLOAD_IDENTITY_AUDIENCE=acme.audience",
				"TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL=terraform@iam.google.com",
				"TFC_GCP_WORKLOAD_PROVIDER_NAME=projects/123/locations/global/workloadIdentityPools/pool-123/providers/provider-123",
			},
			phase: run.PlanPhase,
			want: func(t *testing.T, workdir string, envs []string) {
				want := fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s/gcp_config", workdir)
				assert.Contains(t, envs, want)
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
		})
	}
}
