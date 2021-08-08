package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

const (
	EnvironmentVariablePrefix = "OTS_"
)

// SetFlagsFromEnvVariables overrides flag values with environment variables.
func SetFlagsFromEnvVariables(fs *pflag.FlagSet) {
	fs.VisitAll(func(f *pflag.Flag) {
		envVar := flagToEnvVarName(f)
		if val, present := os.LookupEnv(envVar); present {
			fs.Set(f.Name, val)
		}
	})
}

func flagToEnvVarName(f *pflag.Flag) string {
	return fmt.Sprintf("%s%s", EnvironmentVariablePrefix, strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
