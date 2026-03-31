//go:build darwin

package winds

import (
	"net"
)

const (
	pingProtocolNetworkIPv4 = "udp4"
	pingProtocolNetworkIPv6 = "udp6"
)

type (
	pingTargetAddressType = net.UDPAddr
)
