package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	EnvironmentVariablePrefix = "OTF_"
)

// SetFlagsFromEnvVariables overrides flag values with environment variables.
func SetFlagsFromEnvVariables(fs *pflag.FlagSet) {
	fs.VisitAll(func(f *pflag.Flag) {
		envVar := flagToEnvVarName(f)
		if val, present := os.LookupEnv(envVar); present {
			fs.Set(f.Name, val)
			return
		}

		// Do not look for _FILE if the application is already expecting a file.
		if strings.HasSuffix(envVar, "_FILE") {
			return
		}

		val, present := os.LookupEnv(envVar + "_FILE")
		if present {
			value, err := os.ReadFile(val)
			if err != nil {
				PrintError(errors.Wrapf(err, "failed to read file %s", envVar+"_FILE"))
				return
			}

			fs.Set(f.Name, string(value))
		}
	})
}

func flagToEnvVarName(f *pflag.Flag) string {
	return fmt.Sprintf("%s%s", EnvironmentVariablePrefix, strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
