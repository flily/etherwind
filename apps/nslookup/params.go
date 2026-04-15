package nslookup

import (
	"fmt"
	"net"
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
	Port      int
	TCP       bool
	Timeout   time.Duration
	Resolver  *dns.Resolver
}

func NewDefaultParams() *Params {
	p := &Params{
		QueryType: []dns.Type{
			dns.TypeA,
			dns.TypeAAAA,
		},
		Server:   "",
		Port:     53,
		TCP:      false,
		Timeout:  2 * time.Second,
		Resolver: nil,
	}

	return p
}

func (p *Params) ReloadResolver() error {
	// assume Server field is already validated
	if p.Server == "" {
		nameServers, err := dns.ParseResolvConf(dns.DefaultSystemResolverConfigurePath)
		if err != nil {
			return fmt.Errorf("failed to parse resolv.conf: %s", err)
		}

		endpoints := make([]dns.Endpoint, 0, len(nameServers.Nameservers))
		for _, ip := range nameServers.Nameservers {
			if p.TCP {
				endpoints = append(endpoints, dns.NewTCPEndpoint(ip, p.Port))
			} else {
				endpoints = append(endpoints, dns.NewUDPEndpoint(ip, p.Port))
			}
		}

		p.Resolver = dns.NewResolver(endpoints)

	} else {
		ip := net.ParseIP(p.Server)
		if ip == nil {
			return fmt.Errorf("invalid server IP: %s", p.Server)
		}

		endpoints := []dns.Endpoint{nil}
		if p.TCP {
			fmt.Printf("VC\n")
			endpoints[0] = dns.NewTCPEndpoint(ip, p.Port)

		} else {
			endpoints[0] = dns.NewUDPEndpoint(ip, p.Port)
		}

		p.Resolver.Reload(endpoints)
	}

	fmt.Printf("resolvers: %s\n", p.Resolver.Endpoints)
	return nil
}
