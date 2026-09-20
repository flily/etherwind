package ping

import (
	"fmt"
	"net"
)

func runTuiPing(params *Params, records []*PingRecord) {
	// not implemented
}

func MainTui(params *Params) {
	records := make([]*PingRecord, 0, len(params.Target))
	for _, target := range params.Target {
		ip := net.ParseIP(target)
		if ip == nil {
			fmt.Printf("Invalid IP address: %s\n", target)
			return
		}
		records = append(records, NewPingRecordWithIP(ip, DefaultPingRecordCapacity))
	}

	runTuiPing(params, records)
}
