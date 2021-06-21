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

// Each flag can also be set with an env variable whose name starts with `OTS_`.
func SetFlagsFromEnvVariables(fs *pflag.FlagSet) {
	fs.VisitAll(func(f *pflag.Flag) {
		envVar := flagToEnvVarName(f)
		if val, present := os.LookupEnv(envVar); present {
			fs.Set(f.Name, val)
		}
	})
}

// Unset env vars prefixed with `OTS_`
func UnsetEtokVars() {
	for _, kv := range os.Environ() {
		parts := strings.Split(kv, "=")
		if strings.HasPrefix(parts[0], EnvironmentVariablePrefix) {
			if err := os.Unsetenv(parts[0]); err != nil {
				panic(err.Error())
			}
		}
	}
}

func flagToEnvVarName(f *pflag.Flag) string {
	return fmt.Sprintf("%s%s", EnvironmentVariablePrefix, strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
