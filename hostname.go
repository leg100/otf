package otf

import (
	"fmt"
	"net"
	"strings"
)

type HostnameService interface {
	Hostname() string
}

// SetHostname sets the system hostname. If hostname is provided it'll use
// that; otherwise the listen address is used (which should be that used by the
// otfd daemon).
func SetHostname(hostname string, listen *net.TCPAddr) (string, error) {
	if hostname != "" {
		return hostname, nil
	} else if listen != nil {
		// If ip is unspecified assume 127.0.0.1 - an IP is used instead of
		// 'localhost' because terraform insists on a dot in the registry hostname.
		if listen.IP.IsUnspecified() {
			return fmt.Sprintf("127.0.0.1:%d", listen.Port), nil
		} else {
			return fmt.Sprintf("%s:%d", listen.IP.String(), listen.Port), nil
		}
	} else {
		return "", fmt.Errorf("neither hostname nor listen adddress specified")
	}
}

// HostnameCredentialEnv returns the environment variable key for an API
// token specific to the given hostname.
func HostnameCredentialEnv(hostname string) string {
	return fmt.Sprintf("TF_TOKEN_%s", strings.ReplaceAll(hostname, ".", "_"))
}
