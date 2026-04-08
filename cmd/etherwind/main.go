package main

import (
	"flag"
	"fmt"

	"github.com/flily/etherwind/apps/nslookup"
	"github.com/flily/etherwind/apps/ping"
)

var commandList = map[string]func(args []string){
	"ping":     ping.Main,
	"nslookup": nslookup.Main,
}

func usage() {
	fmt.Printf("Usage: etherwind <command> [args]\n")
	fmt.Printf("\n")
	fmt.Printf("Available commands:\n")
	for cmd := range commandList {
		fmt.Printf("  %s\n", cmd)
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		return
	}

	cmd := args[0]
	entry, ok := commandList[cmd]
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		return
	}

	entry(args[1:])
}
