//go:build !darwin

package winds

import (
	"net"
)

func pingerGetNetwork(network string) string {
	switch network {
	case NetworkIPv4:
		return "ip4:icmp"
	case NetworkIPv6:
		return "ip6:ipv6-icmp"
	default:
		panic("invalid network type " + network)
	}
}

func pingerMakeAddress(address net.IP) net.Addr {
	return &net.IPAddr{IP: address}
}
