package nslookup

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"slices"

	"github.com/flily/etherwind/common/dns"
)

const (
	Prompt = "> "
)

type Command struct {
	Command     string
	OptionName  string
	OptionValue string
	Name        string
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// COMMAND  OPTION=VALUE  NAME
// ^      ^ ^      ^    ^ ^
// 0      1 2      3    4 5
func ParseCommand(line string) Command {
	cmd := Command{}

	content := []rune(line)
	buf := make([]rune, 0, len(content))

	state := 0
	for i := 0; i < len(content); i++ {
		switch state {
		case 0:
			if isSpace(content[i]) {
				cmd.Command = string(buf)
				buf = buf[:0]
				state = 1

			} else {
				buf = append(buf, content[i])
			}

		case 1:
			if !isSpace(content[i]) {
				state = 2
				buf = append(buf, content[i])
			}

		case 2:
			if content[i] == '=' {
				state = 3
				cmd.OptionName = string(buf)
				buf = buf[:0]

			} else if isSpace(content[i]) {
				state = 4
				cmd.OptionName = string(buf)
				buf = buf[:0]

			} else {
				buf = append(buf, content[i])
			}

		case 3:
			if isSpace(content[i]) {
				state = 4
				cmd.OptionValue = string(buf)
				buf = buf[:0]

			} else {
				buf = append(buf, content[i])
			}

		case 4:
			if !isSpace(content[i]) {
				state = 5
				buf = append(buf, content[i])
			}

		case 5:
			buf = append(buf, content[i])
		}
	}

	lastValue := string(buf)
	switch state {
	case 0:
		if slices.Contains(SupportedCommands, lastValue) {
			cmd.Command = lastValue
		} else {
			cmd.Name = lastValue
		}

	case 2:
		cmd.OptionName = lastValue

	case 3:
		cmd.OptionValue = lastValue

	case 4, 5:
		cmd.Name = lastValue
	}

	return cmd
}

func runQueryCommand(params *Params, name string) {
	var ns *dns.Resolver
	if len(params.Server) <= 0 {
		rsv, err := dns.NewDefaultResolver()
		if err != nil {
			fmt.Printf("failed to load default nameserver")
			return
		}
		ns = rsv

	} else {
		ns = dns.NewResolver([]dns.Endpoint{
			dns.NewUDPEndpoint(net.ParseIP(params.Server), dns.DNSDefaultPort),
		})
	}

	result, from, err := ns.QueryA(context.Background(), name)
	if err != nil {
		fmt.Printf("failed to lookup IP: %v\n", err)
		return
	}

	fmt.Printf("Server:\t\t%s\n", from.Address())
	fmt.Printf("Address:\t%s\n", from.FullAddress())
	fmt.Printf("\n")

	for _, ip := range result {
		fmt.Printf("Name:\t%s\n", name)
		fmt.Printf("Address: %s\n", ip.String())
		fmt.Printf("\n")
	}
}

func updateParams(params *Params, cmd Command, r *dns.Resolver) bool {
	switch cmd.Command {
	case "exit":
		return true

	case "server":
		params.Server = cmd.OptionName
		server := net.ParseIP(cmd.OptionName)
		if server == nil {
			fmt.Printf("invalid server IP: %s\n", cmd.OptionName)
			return false
		}

		loadedServer := r.Reload([]dns.Endpoint{
			dns.NewUDPEndpoint(server, dns.DNSDefaultPort),
		})
		fmt.Printf("server:\t\t%s\n", server)
		fmt.Printf("Address:\t%s\n", loadedServer[0])

	}

	return false
}

func MainClassicalInteractiveLoop(finished chan<- struct{}) {
	r, err := dns.NewDefaultResolver()
	if err != nil {
		fmt.Printf("failed to load default nameserver: %v\n", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	params := NewDefaultParams()

	for {
		fmt.Print(Prompt)
		line, _, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			fmt.Printf("failed to read input: %v\n", err)
			continue
		}

		cmd := ParseCommand(string(line))
		if len(cmd.Command) <= 0 {
			runQueryCommand(params, cmd.Name)

		} else {
			exit := updateParams(params, cmd, r)
			if exit {
				break
			}
		}
	}

	finished <- struct{}{}
}

func MainClassicalInteractive() {
	finished := make(chan struct{})
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go MainClassicalInteractiveLoop(finished)

	select {
	case <-finished:
	case <-ctx.Done():
	}
}

func MainClassicalNonInteractive(args []string) {
	target := ""
	if len(args) > 0 {
		target = args[0]
	}

	params := NewDefaultParams()
	nameserver := ""
	if len(args) > 1 {
		nameserver = args[len(args)-1]
	}
	params.Server = nameserver

	runQueryCommand(params, target)

}

func MainClassical(args []string) {
	if len(args) > 0 {
		MainClassicalNonInteractive(args)

	} else {
		MainClassicalInteractive()
	}
}
