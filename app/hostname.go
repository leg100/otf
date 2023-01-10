package app

import (
	"net"

	"github.com/leg100/otf"
)

func (a *Application) Hostname() string { return a.hostname }

func (a *Application) SetHostname(hostname string, listen *net.TCPAddr) error {
	hostname, err := otf.SetHostname(hostname, listen)
	if err != nil {
		return err
	}
	a.hostname = hostname
	a.V(0).Info("set system hostname", "hostname", a.hostname)
	return nil
}
