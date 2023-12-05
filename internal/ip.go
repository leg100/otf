package internal

import (
	"net"
)

// GetOutboundIP gets the preferred outbound IP address of this machine.
//
// Credit to: https://stackoverflow.com/a/37382208
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, nil
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP, nil
}
