package internal

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

type (
	// HostnameService is registry of hostnames
	HostnameService interface {
		// Return the OTF hostname.
		Hostname() string
		// Set the OTF hostname.
		SetHostname(string)
		// Return OTF URL with the given path
		URL(path string) string
	}

	hostnameService struct {
		hostname string
	}
)

func NewHostnameService(hostname string) *hostnameService {
	return &hostnameService{hostname}
}

func (s *hostnameService) Hostname() string            { return s.hostname }
func (s *hostnameService) SetHostname(hostname string) { s.hostname = hostname }

func (s *hostnameService) URL(path string) string {
	u := url.URL{
		Scheme: "https",
		Host:   s.Hostname(),
		Path:   path,
	}
	return u.String()
}

// NormalizeAddress takes a host:port and converts it into a host:port
// appropriate for setting as the addressable hostname of otfd, e.g. converting
// 0.0.0.0 to 127.0.0.1.
func NormalizeAddress(addr *net.TCPAddr) string {
	// If ip is unspecified assume 127.0.0.1 - an IP is used instead of
	// 'localhost' because terraform insists on a dot in the registry hostname.
	if addr.IP.IsUnspecified() {
		return fmt.Sprintf("127.0.0.1:%d", addr.Port)
	}
	return fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
}

// CredentialEnvKey returns the environment variable key for an API
// token specific to the given hostname.
func CredentialEnvKey(hostname string) string {
	return fmt.Sprintf("TF_TOKEN_%s", strings.ReplaceAll(hostname, ".", "_"))
}

// CredentialEnv returns a host-specific environment variable credential for
// terraform.
func CredentialEnv(hostname string, token []byte) string {
	return fmt.Sprintf("%s=%s", CredentialEnvKey(hostname), string(token))
}
