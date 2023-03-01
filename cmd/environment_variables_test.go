package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFlagsFromEnvVariables(t *testing.T) {
	t.Run("override flag with env var", func(t *testing.T) {
		fs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		got := fs.String("foo", "default", "")
		t.Setenv("OTF_FOO", "bar")
		require.NoError(t, SetFlagsFromEnvVariables(fs))
		require.NoError(t, fs.Parse(nil))
		assert.Equal(t, "bar", *got)
	})
	t.Run("override flag with env var file", func(t *testing.T) {
		fs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		got := fs.String("foo", "default", "")
		t.Setenv("OTF_FOO_FILE", "./testdata/otf_foo_file")
		require.NoError(t, SetFlagsFromEnvVariables(fs))
		require.NoError(t, fs.Parse(nil))
		assert.Equal(t, "big\nmultiline\nsecret\n", *got)
	})
	t.Run("ignore env var for flag ending with _file", func(t *testing.T) {
		fs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		got := fs.String("foo_file", "default", "")
		t.Setenv("OTF_FOO_FILE_FILE", "./testdata/otf_foo_file")
		require.NoError(t, SetFlagsFromEnvVariables(fs))
		require.NoError(t, fs.Parse(nil))
		assert.Equal(t, "default", *got)
	})
	t.Run("override flag with non-existent env var file", func(t *testing.T) {
		fs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		_ = fs.String("foo", "default", "")
		t.Setenv("OTF_FOO_FILE", "./does-not-exist")
		assert.Error(t, SetFlagsFromEnvVariables(fs))
	})
}
