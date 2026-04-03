package ping

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/flily/etherwind/winds"
)

func MainClassical(params *Params) {
	addr := net.ParseIP(params.Target[0])
	if addr == nil {
		fmt.Printf("Invalid IP address: %s\n", params.Target[0])
		return
	}

	network := winds.NetworkIPv4
	if addr.To4() == nil {
		network = winds.NetworkIPv6
	}

	pinger, err := winds.NewPinger(network)
	if err != nil {
		errMessage := rootError(err)
		if isPermissionDenied(err) {
			fmt.Printf("No permission to start ping: %s\n", errMessage.Error())
			if runtime.GOOS == "linux" {
				fmt.Printf("Hint: use `sudo setcap cap_net_raw=+ep etherwind` or `sudo` to get root privileges\n")
			}

		} else {
			fmt.Printf("Failed to create pinger: %s\n", errMessage.Error())
		}

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
			if operr, ok := err.(*net.OpError); ok && operr.Timeout() {
				fmt.Printf("Request timeout for icmp_seq %d\n", seq)
				seq++

			} else {
				errMessage := errors.Unwrap(err)
				fmt.Printf("ping: %s\n", errMessage)
			}

			time.Sleep(1 * time.Second)
			continue
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
