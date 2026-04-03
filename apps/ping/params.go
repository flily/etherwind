package ping

import (
	"flag"
	"time"
)

type Params struct {
	Count     int
	Interval  time.Duration
	TTL       int
	Timeout   time.Duration
	Classical bool
	Target    []string
}

func DefaultParams() *Params {
	params := &Params{
		Count:     -1,
		Interval:  1 * time.Second,
		TTL:       64,
		Timeout:   1 * time.Second,
		Classical: true,
		Target:    []string{},
	}

	return params
}

func ParseParams(name string, args []string) (*flag.FlagSet, *Params) {
	params := DefaultParams()
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	set.IntVar(&params.Count, "count", params.Count,
		"Stop after sending (and receiving) count ECHO_RESPONSE packets.")
	set.DurationVar(&params.Interval, "interval", params.Interval,
		"Wait INTERVAL between sending each packet. The default is to wait for a response before sending the next packet.")
	set.IntVar(&params.TTL, "ttl", params.TTL,
		"Time to live in the IP header.")
	set.DurationVar(&params.Timeout, "timeout", params.Timeout,
		"Time to wait for a response.")
	set.BoolVar(&params.Classical, "classical", params.Classical,
		"Output in classical format.")

	_ = set.Parse(args)
	params.Target = set.Args()

	return set, params
}
