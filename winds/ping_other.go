//go:build !darwin

package winds

import (
	"net"
)

const (
	pingProtocolNetworkIPv4 = "ip4:icmp"
	pingProtocolNetworkIPv6 = "ip6:ipv6-icmp"
)

type (
	pingTargetAddressType = net.IPAddr
)
