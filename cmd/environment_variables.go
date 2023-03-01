package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	EnvironmentVariablePrefix     = "OTF_"
	EnvironmentVariableFileSuffix = "_FILE"
)

// SetFlagsFromEnvVariables overrides flag values with environment variables.
func SetFlagsFromEnvVariables(fs *pflag.FlagSet) error {
	fileEnvs := make(map[string]*pflag.Flag, fs.NFlag())
	fs.VisitAll(func(f *pflag.Flag) {
		envVar := flagToEnvVarName(f)
		if val, present := os.LookupEnv(envVar); present {
			fs.Set(f.Name, val)
			return
		}

		// Do not look for _FILE if the application is already expecting a file.
		if strings.HasSuffix(envVar, EnvironmentVariableFileSuffix) {
			return
		}

		if _, present := os.LookupEnv(envVar + EnvironmentVariableFileSuffix); present {
			fileEnvs[envVar+EnvironmentVariableFileSuffix] = f
		}
	})

	for envVar, f := range fileEnvs {
		value, err := os.ReadFile(os.Getenv(envVar))
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", envVar)
		}

		fs.Set(f.Name, string(value))
	}

	return nil
}

func flagToEnvVarName(f *pflag.Flag) string {
	return fmt.Sprintf("%s%s", EnvironmentVariablePrefix, strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
