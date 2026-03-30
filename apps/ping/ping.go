package ping

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/flily/etherwind/winds"
)

func rootError(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}

func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, os.ErrPermission) {
		fmt.Printf("os.ErrPermission: %s\n", err)
		return true
	}

	if errors.Is(err, syscall.EPERM) {
		fmt.Printf("syscall.EPERM: %s\n", err)
		return true
	}

	if errors.Is(err, syscall.EACCES) {
		fmt.Printf("syscall.EACCES: %s\n", err)
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "permission denied")
}

func createPinger(addr net.IP) (*winds.Pinger, string, error) {
	networkCandidates := []string{"udp4", "ip4:icmp"}
	if addr.To4() == nil {
		networkCandidates = []string{"udp6", "ip6:ipv6-icmp"}
	}

	var lastErr error
	for _, network := range networkCandidates {
		pinger, err := winds.NewPinger(network)
		if err == nil {
			return pinger, network, nil
		}

		lastErr = err
		fmt.Printf("ERROR ON '%s': %s\n", network, err)
		if !isPermissionDenied(err) {
			break
		}
	}

	return nil, "", lastErr
}

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

	pinger, network, err := createPinger(addr)
	if err != nil {
		errMessage := rootError(err)
		if isPermissionDenied(err) {
			fmt.Printf("No permission to start ping: %s\n", errMessage.Error())
			fmt.Printf("Hint: use `sudo setcap cap_net_raw=+ep etherwind` or `sudo` to get root privileges\n")
			return
		}

		fmt.Printf("Error creating pinger: %s\n", errMessage.Error())
		return
	}
	defer func() {
		_ = pinger.Close()
	}()

	if addr.To4() == nil {
		fmt.Printf("ipv6\n")
	}
	fmt.Printf("Using network mode: %s\n", network)

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
