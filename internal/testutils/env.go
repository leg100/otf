package testutils

import (
	"os"
	"testing"
)

func SkipIfEnvUnspecified(t *testing.T, env string) {
	if _, ok := os.LookupEnv(env); !ok {
		t.Skip("Export valid " + env + " before running this test")
	}
}
