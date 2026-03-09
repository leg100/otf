package internal

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/leg100/otf/internal/logr"
)

type HostnameService struct {
	hostname        string
	webhookHostname string
	localPort       int
}

func NewHostnameService(
	logger logr.Logger,
	hostname,
	webhookHostname string,
	listenAddress *net.TCPAddr,
) *HostnameService {
	// Unless an explicit hostname has been set, set it to the listening address
	// of the http server.
	if hostname == "" {
		hostname = normalizeAddress(listenAddress)
	}
	// Unless an explicit webhook hostname has been set, set it to whatever the
	// hostname is set to.
	if webhookHostname == "" {
		webhookHostname = hostname
	}

	logger.V(0).Info("set system hostname", "hostname", hostname)
	logger.V(0).Info("set webhook hostname", "webhook_hostname", webhookHostname)

	return &HostnameService{
		hostname:        hostname,
		webhookHostname: webhookHostname,
		localPort:       listenAddress.Port,
	}
}

func (s *HostnameService) Hostname() string        { return s.hostname }
func (s *HostnameService) WebhookHostname() string { return s.webhookHostname }

func (s *HostnameService) URL(path string) string {
	u := url.URL{
		Scheme: "https",
		Host:   s.Hostname(),
		Path:   path,
	}
	return u.String()
}

func (s *HostnameService) WebhookURL(path string) string {
	u := url.URL{
		Scheme: "https",
		Host:   s.WebhookHostname(),
		Path:   path,
	}
	return u.String()
}

// LocalURL returns an absolute URL for a path with the host set to
// localhost. Useful for testing purposes where hostname might be set to
// something that isn't routable from the local machine.
func (s *HostnameService) LocalURL(path string) string {
	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("localhost:%d", s.localPort),
		Path:   path,
	}
	return u.String()
}

// normalizeAddress takes a host:port and converts it into a host:port
// appropriate for setting as the addressable hostname of otfd, e.g. converting
// 0.0.0.0 to 127.0.0.1.
func normalizeAddress(addr *net.TCPAddr) string {
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
