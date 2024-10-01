package internal

import (
	"net"
	"net/netip"
)

// GetOutboundIP gets the preferred outbound IP address of this machine.
//
// Credit to: https://stackoverflow.com/a/37382208
func GetOutboundIP() (netip.Addr, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return netip.Addr{}, nil
	}
	defer conn.Close()

	return netip.ParseAddr(conn.LocalAddr().String())
}
