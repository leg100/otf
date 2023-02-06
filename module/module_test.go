package module

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

func TestSortedModuleVersions(t *testing.T) {
	org := NewTestOrganization(t)
	mod := NewTestModule(org)

	v0_1 := NewTestModuleVersion(mod, "0.1", ModuleVersionStatusOk)
	v0_2 := NewTestModuleVersion(mod, "0.2", ModuleVersionStatusOk)
	v0_3 := NewTestModuleVersion(mod, "0.3", ModuleVersionStatusOk)
	v0_4 := NewTestModuleVersion(mod, "0.4", ModuleVersionStatusPending)

	t.Run("add", func(t *testing.T) {
		l := SortedModuleVersions{}
		l = l.add(v0_1)
		l = l.add(v0_4)
		l = l.add(v0_3)
		l = l.add(v0_2)

		assert.Equal(t, SortedModuleVersions{v0_1, v0_2, v0_3, v0_4}, l)
	})

	t.Run("latest", func(t *testing.T) {
		assert.Nil(t, SortedModuleVersions{}.latest())
		assert.Equal(t, v0_1, SortedModuleVersions{v0_1}.latest())
		assert.Equal(t, v0_2, SortedModuleVersions{v0_1, v0_2}.latest())
		assert.Equal(t, v0_3, SortedModuleVersions{v0_1, v0_2, v0_3}.latest())
		assert.Equal(t, v0_3, SortedModuleVersions{v0_1, v0_2, v0_3, v0_4}.latest())
	})
}
