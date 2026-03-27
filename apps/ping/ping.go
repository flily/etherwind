package ping

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/flily/etherwind/winds"
)

func Main(args []string) {
	if len(args) < 1 {
		fmt.Printf("Usage: etherwind <command> [args]\n")
		return
	}

	addr := net.ParseIP(args[0])
	if addr == nil {
		fmt.Printf("Invalid IP address: %s\n", args[0])
		return
	}

	network := "udp4"
	if addr.To4() == nil {
		network = "udp6"
		fmt.Printf("ipv6\n")
	}

	pinger, err := winds.NewPinger(network)
	if err != nil {
		fmt.Printf("Error creating pinger: %s\n", err.Error())
		return
	}
	defer func() {
		_ = pinger.Close()
	}()

	payloadBase := winds.DefaultPingPayloadBase
	fmt.Printf("PING %s: %d data bytes\n", addr, len(payloadBase)+16)

	seq := 1
	id := os.Getpid() & 0xffff
	for {
		result, err := pinger.Ping(addr, id, seq, payloadBase)
		if err != nil {
			fmt.Printf("Error pinging %s: %s\n", addr, err.Error())
			return
		}

		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.6f ms\n",
			len(result.Raw),
			addr,
			result.Seq,
			result.TTL,
			float64(result.Duration)/float64(time.Millisecond),
		)
		seq++

		time.Sleep(1 * time.Second)
	}
}
