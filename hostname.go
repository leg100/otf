package otf

import (
	"fmt"
	"net"
)

type HostnameService struct {
	hostname string
}

func (s *HostnameService) SetHostname(hostname string, ln net.Listener) error {
	ln.Addr().(*net.TCPAddr)
	if hostname != "" {
		s.hostname = hostname
		return nil
	}
	if listen != "" {
		host, port, err := net.SplitHostPort(listen)
		if err != nil {
			return err
		}
		if host == "" {
			host = "127.0.0.1"
		}
		s.hostname = host + ":" + port
		return nil
	}
	return fmt.Errorf("neither hostname nor listen adddress specified")
}
