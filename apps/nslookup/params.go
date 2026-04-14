package nslookup

import (
	"time"

	"github.com/flily/etherwind/common/dns"
)

var (
	SupportedCommands = []string{
		"host",
		"server",
		"lserver",
		"finger",
		"ls",
		"view",
		"help",
		"?",
		"exit",
		"set",
	}

	SupportedQueryTypes = []string{
		"A",
		"AAAA",
		"CNAME",
		"MX",
		"NS",
	}
)

type Params struct {
	QueryType []dns.Type
	Server    string
	TCP       bool
	Timeout   time.Duration
}

func NewDefaultParams() *Params {
	p := &Params{
		QueryType: []dns.Type{
			dns.TypeA,
			dns.TypeAAAA,
		},
		Server:  "",
		TCP:     false,
		Timeout: 2 * time.Second,
	}

	return p
}
