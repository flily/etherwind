package nslookup

import (
	"context"
	"fmt"
	"net"

	"github.com/flily/etherwind/common/resolver"
)

func MainClassicalInteractive() {

}

func MainClassicalNonInteractive(args []string) {
	fmt.Printf("args: %s\n", args)
	target := ""
	if len(args) > 0 {
		target = args[0]
	}

	nameserver := ""
	if len(args) > 1 {
		nameserver = args[len(args)-1]
	}

	fmt.Printf("nameserver=[%s]]n", nameserver)
	var ns *resolver.Resolver
	if len(nameserver) <= 0 {
		fmt.Printf("------\n")
		rsv, err := resolver.NewDefaultResolver()
		if err != nil {
			fmt.Printf("failed to load default nameserver")
			return
		}
		ns = rsv

	} else {
		fmt.Printf("=======\n")
		ns = resolver.NewResolver([]net.Addr{
			&net.UDPAddr{
				IP:   net.ParseIP(nameserver),
				Port: resolver.DNSDefaultPort,
			},
		})
	}

	result, from, err := ns.LookupIP(context.Background(), target)
	if err != nil {
		fmt.Printf("failed to lookup IP: %v\n", err)
		return
	}

	fmt.Printf("Server:\t\t%s\n", from.String())
	fmt.Printf("Address:\t%s/%s\n", from.String(), from.Network())
	fmt.Printf("\n")

	for _, ip := range result {
		fmt.Printf("Name:\t%s\n", target)
		fmt.Printf("Address: %s\n", ip.String())
		fmt.Printf("\n")
	}
}

func MainClassical(args []string) {
	if len(args) > 0 {
		MainClassicalNonInteractive(args)

	} else {
		MainClassicalInteractive()
	}
}
