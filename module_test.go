package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextModuleStatus(t *testing.T) {
	tests := []struct {
		name          string
		current       ModuleStatus
		versionStatus ModuleVersionStatus
		want          ModuleStatus
	}{
		{
			name:          "pending -> setup_complete",
			current:       ModuleStatusPending,
			versionStatus: ModuleVersionStatusOk,
			want:          ModuleStatusSetupComplete,
		},
		{
			name:          "pending -> setup_failed",
			current:       ModuleStatusPending,
			versionStatus: ModuleVersionStatusRegIngressFailed,
			want:          ModuleStatusSetupFailed,
		},
		{
			name:          "no version tags -> setup_complete",
			current:       ModuleStatusNoVersionTags,
			versionStatus: ModuleVersionStatusOk,
			want:          ModuleStatusSetupComplete,
		},
		{
			name:          "setup_complete -> setup_complete",
			current:       ModuleStatusSetupComplete,
			versionStatus: ModuleVersionStatusOk,
			want:          ModuleStatusSetupComplete,
		},
		{
			name:          "setup_complete -> setup_complete, despite bad module version",
			current:       ModuleStatusSetupComplete,
			versionStatus: ModuleVersionStatusRegIngressFailed,
			want:          ModuleStatusSetupComplete,
		},
		{
			name:          "setup_failed -> setup_complete",
			current:       ModuleStatusSetupFailed,
			versionStatus: ModuleVersionStatusOk,
			want:          ModuleStatusSetupComplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextModuleStatus(tt.current, tt.versionStatus)
			assert.Equal(t, tt.want, got)
		})
	}
}
