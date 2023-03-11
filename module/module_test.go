package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	modver1 := ModuleVersion{Version: "v1", Status: ModuleVersionStatusOK}
	modver2 := ModuleVersion{Version: "v2", Status: ModuleVersionStatusOK}
	modver3 := ModuleVersion{Version: "v3", Status: ModuleVersionStatusPending}
	mod := &Module{Versions: []ModuleVersion{modver3, modver2, modver1}}

	t.Run("latest", func(t *testing.T) {
		assert.Equal(t, &modver2, mod.Latest())
	})

	t.Run("available", func(t *testing.T) {
		assert.Equal(t, []ModuleVersion{modver2, modver1}, mod.AvailableVersions())
	})

	t.Run("version", func(t *testing.T) {
		assert.Equal(t, &modver2, mod.Version("v2"))
	})
}
