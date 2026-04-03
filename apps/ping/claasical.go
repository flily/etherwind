package ping

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/flily/etherwind/winds"
)

func runClassicalPing(params *Params, records *TimeRecord, finished chan struct{}) {
	addr := net.ParseIP(params.Target[0])
	if addr == nil {
		fmt.Printf("Invalid IP address: %s\n", params.Target[0])
		finished <- struct{}{}
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

		finished <- struct{}{}
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

		pingTimeMs := float64(result.Duration) / float64(time.Millisecond)
		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.6f ms\n",
			len(result.Raw),
			addr,
			result.Seq,
			result.TTL,
			pingTimeMs,
		)
		records.Add(pingTimeMs)
		seq++

		time.Sleep(1 * time.Second)
	}

	finished <- struct{}{}
}

func MainClassical(params *Params) {
	finished := make(chan struct{})
	records := NewTimeRecords(3600)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go runClassicalPing(params, records, finished)

	select {
	case <-finished:

	case <-ctx.Done():
		fmt.Printf("\n--- %s ping statistics ---\n", params.Target[0])
		fmt.Printf("%d packets transmitted, %d received, %.2f%% packet loss\n",
			len(records.Records), len(records.Records), 100.0*(1.0-float64(len(records.Records))/float64(len(records.Records))))
		fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n",
			records.Min(), records.Average(), records.Max(), records.StandardDeviation(),
		)
	}
}
