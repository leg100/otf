package otf

import (
	"fmt"
	"net"
	"strings"

	"github.com/go-logr/logr"
)

type (
	// HostnameService provides the OTF user-facing hostname.
	HostnameService interface {
		Hostname() string
	}

	hostnameService struct {
		logr.Logger

		hostname string
	}
)

func NewHostnameService(logger logr.Logger) *hostnameService {
	return &hostnameService{
		Logger: logger,
	}
}

func (a *hostnameService) Hostname() string { return a.hostname }

func (a *hostnameService) SetHostname(hostname string, listen *net.TCPAddr) error {
	hostname, err := setHostname(hostname, listen)
	if err != nil {
		return err
	}
	a.hostname = hostname

	a.V(0).Info("set system hostname", "hostname", a.hostname)
	return nil
}

// setHostname sets the system hostname. If hostname is provided it'll use
// that; otherwise the listen address is used (which should be that used by the
// otfd daemon).
func setHostname(hostname string, listen *net.TCPAddr) (string, error) {
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
