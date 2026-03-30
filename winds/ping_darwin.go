//go:build darwin

package winds

import (
	"net"
)

func pingerGetNetwork(network string) string {
	switch network {
	case NetworkIPv4:
		return "udp4"
	case NetworkIPv6:
		return "udp6"
	default:
		panic("invalid network type " + network)
	}
}

func pingerMakeAddress(address net.IP) net.Addr {
	return &net.UDPAddr{IP: address}
}
