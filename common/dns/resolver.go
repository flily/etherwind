package dns

import (
	"net"

	"github.com/flily/etherwind/common/resolver"
	"golang.org/x/net/dns/dnsmessage"
)

type Resolver struct {
	Endpoints []Endpoint
	targets   []*Client
	index     int
}

func NewResolver(endpoints []Endpoint) *Resolver {
	r := &Resolver{}
	r.Reload(endpoints)
	return r
}

func NewDefaultResolver() (*Resolver, error) {
	conf, err := resolver.ParseResolvConf(resolver.DefaultSystemResolverConfigurePath)
	if err != nil {
		return nil, err
	}

	endpoints := make([]Endpoint, 0, len(conf.Nameservers))
	for _, ip := range conf.Nameservers {
		ep := NewUDPEndpoint(ip, 53)
		endpoints = append(endpoints, ep)
	}

	return NewResolver(endpoints), nil
}

func (r *Resolver) Reload(endpoints []Endpoint) []Endpoint {
	r.Endpoints = endpoints
	r.targets = make([]*Client, len(endpoints))
	r.index = 0

	return r.Endpoints
}

func (r *Resolver) getClient() (*Client, error) {
	client := r.targets[r.index]
	if client == nil {
		endpoint := r.Endpoints[r.index]
		client = NewClient(endpoint)
		err := client.Dial()
		if err != nil {
			return nil, err
		}

		r.targets[r.index] = client
	}

	r.index = (r.index + 1) % len(r.Endpoints)
	return client, nil
}

func (r *Resolver) QueryRaw(t Type, name string) (*Message, error) {
	client, err := r.getClient()
	if err != nil {
		return nil, err
	}

	response, err := client.Query(t, name)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *Resolver) QueryA(name string) ([]net.IP, error) {
	response, err := r.QueryRaw(TypeA, name)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, 0, 16)
	for _, answer := range response.Answers {
		if answer.Header.Type == TypeA {
			ips = append(ips, answer.Body.(*dnsmessage.AResource).A[:])
		}
	}

	return ips, nil
}
