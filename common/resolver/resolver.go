package resolver

import (
	"context"
	"net"
	"slices"
)

const (
	DNSDefaultPort = 53
)

type ConfigureType int

const (
	ConfigureTypeCustom ConfigureType = iota
	ConfigureTypeResolvConf
)

type Resolver struct {
	from        string
	fromType    ConfigureType
	NameServers []net.Addr
	queryCount  int
	resolvers   []*net.Resolver
}

func makeDNSDialer(addr net.Addr) func(ctx context.Context, network, address string) (net.Conn, error) {
	switch a := addr.(type) {
	case *net.IPAddr:
		return func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.DialUDP("udp", nil, &net.UDPAddr{
				IP:   a.IP,
				Port: DNSDefaultPort,
			})
		}

	case *net.UDPAddr:
		return func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.DialUDP("udp", nil, a)
		}

	case *net.TCPAddr:
		return func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.DialTCP("tcp", nil, a)
		}

	case *net.UnixAddr:
		return func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.DialUnix("unix", nil, a)
		}

	default:
		panic("unsupported address type")
	}
}

func ToAddrs[T net.Addr](ips []T) []net.Addr {
	addrs := make([]net.Addr, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, ip)
	}

	return addrs
}

func newResolver(servers []net.Addr) *Resolver {
	ns := make([]*net.Resolver, 0, len(servers))
	for _, s := range servers {
		nameserver := &net.Resolver{
			PreferGo: true,
			Dial:     makeDNSDialer(s),
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

	r := newResolver(ToAddrs(conf.MakeDefaultUDPEndpoints(DNSDefaultPort)))
	r.from = path
	r.fromType = ConfigureTypeResolvConf

	return r, nil
}

func NewDefaultResolver() (*Resolver, error) {
	return NewResolverFrom(DefaultSystemResolverConfigurePath)
}

func NewResolver(nameServers []net.Addr) *Resolver {
	r := newResolver(nameServers)
	r.from = ""
	r.fromType = ConfigureTypeCustom

	return r
}

func NewResolverFromIP(addresses []net.IP, port int) *Resolver {
	addrs := make([]net.Addr, 0, len(addresses))
	for _, ip := range addresses {
		addr := &net.UDPAddr{
			IP:   ip,
			Port: port,
		}
		addrs = append(addrs, addr)
	}

	return NewResolver(addrs)
}

func (r *Resolver) From() string {
	return r.from
}

func (r *Resolver) FromType() ConfigureType {
	return r.fromType
}

func (r *Resolver) LookupIP(ctx context.Context, name string) ([]net.IP, net.Addr, error) {
	index := r.queryCount % len(r.NameServers)
	addr := r.NameServers[index]
	ns := r.resolvers[index]
	r.queryCount++

	ips, err := ns.LookupIP(ctx, "ip", name)
	return ips, addr, err
}
