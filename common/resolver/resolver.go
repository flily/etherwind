package resolver

import (
	"context"
	"net"
	"slices"
)

const (
	DNSDefsultPort = 53
)

type ConfigureType int

const (
	ConfigureTypeCustom ConfigureType = iota
	ConfigureTypeResolvConf
)

type Resolver struct {
	from        string
	fromType    ConfigureType
	NameServers []net.IP
	queryCount  int
	resolvers   []*net.Resolver
}

func newResolver(servers []net.IP) *Resolver {
	ns := make([]*net.Resolver, 0, len(servers))
	for _, s := range servers {
		nameserver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.DialUDP("udp", nil, &net.UDPAddr{
					IP:   s,
					Port: DNSDefsultPort,
				})
			},
		}

		ns = append(ns, nameserver)
	}

	r := &Resolver{
		NameServers: slices.Clone(servers),
		queryCount:  0,
		resolvers:   ns,
	}

	return r
}

func NewResolverFrom(path string) (*Resolver, error) {
	conf, err := ParseResolvConf(path)
	if err != nil {
		return nil, err
	}

	r := newResolver(conf.Nameservers)
	r.from = path
	r.fromType = ConfigureTypeResolvConf

	return r, nil
}

func NewDefaultResolver() (*Resolver, error) {
	return NewResolverFrom(DefaultSystemResolverConfigurePath)
}

func NewResolver(nameServers []net.IP) *Resolver {
	r := newResolver(nameServers)
	r.from = ""
	r.fromType = ConfigureTypeCustom

	return r
}

func (r *Resolver) From() string {
	return r.from
}

func (r *Resolver) FromType() ConfigureType {
	return r.fromType
}

func (r *Resolver) LookupIP(ctx context.Context, name string) ([]net.IP, error) {
	ns := r.resolvers[r.queryCount%len(r.NameServers)]
	r.queryCount++

	return ns.LookupIP(ctx, "ip", name)
}
