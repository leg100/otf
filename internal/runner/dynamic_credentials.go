package runner

type dynamicCredentialsEnabler interface {
	enableDynamicCredentials(envs []string, workdir workdir, token []byte) []string
}
