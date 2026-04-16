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
	"strconv"

	"github.com/flily/etherwind/common/dns"
	"golang.org/x/net/dns/dnsmessage"
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

func showAnswers(answers []dns.Resource) {
	keys := make([]string, 0, len(answers))
	for _, answer := range answers {
		name := answer.Header.Name
		key := dns.ResourceKey(answer)
		if slices.Contains(keys, key) {
			continue
		}

		keys = append(keys, key)

		switch answer.Header.Type {
		case dns.TypeA:
			ans := answer.Body.(*dnsmessage.AResource)
			fmt.Printf("Name:\t%s\n", name)
			fmt.Printf("Address:\t%s\n", net.IP(ans.A[:]).String())

		case dns.TypeAAAA:
			ans := answer.Body.(*dnsmessage.AAAAResource)
			fmt.Printf("Name:\t%s\n", name)
			fmt.Printf("Address:\t%s\n", net.IP(ans.AAAA[:]).String())

		case dns.TypeCNAME:
			ans := answer.Body.(*dnsmessage.CNAMEResource)
			fmt.Printf("%s\tcanonical name = %s\n", name, ans.CNAME)

		case dns.TypeSOA:
			ans := answer.Body.(*dnsmessage.SOAResource)
			fmt.Printf("%s\n", answer.Header.Name)
			fmt.Printf("\torigin = %s\n", ans.NS)
			fmt.Printf("\tmail addr = %s\n", ans.MBox)
			fmt.Printf("\tserial = %d\n", ans.Serial)
			fmt.Printf("\trefresh = %d\n", ans.Refresh)
			fmt.Printf("\tretry = %d\n", ans.Retry)
			fmt.Printf("\texpire = %d\n", ans.Expire)
			fmt.Printf("\tminimum = %d\n", ans.MinTTL)
		}
	}
}

func runQueryCommand(params *Params, name string) {
	ns := params.Resolver

	var from dns.Endpoint
	answers := make([]*dns.Message, 0, len(params.QueryType))
	for _, queryType := range params.QueryType {
		result, f, err := ns.QueryRaw(context.Background(), queryType, name)
		if err != nil {
			fmt.Printf("failed to lookup IP: %v\n", err)
			return
		}

		answers = append(answers, result)
		from = f
	}

	fmt.Printf("Server:\t\t%s\n", from.Address())
	fmt.Printf("Address:\t%s\n", from.FullAddress())
	fmt.Printf("\n")

	result := dns.MergeAnswers(answers...)

	fmt.Printf("Non-authoritative answer:\n")
	showAnswers(result.Answers)

	if slices.Contains(params.QueryType, dns.TypeAAAA) {
		fmt.Printf("\n")
		fmt.Printf("Authoritative answers can be found from:\n")
		showAnswers(result.Authorities)
	}
}

func updateParamsSet(params *Params, cmd Command) {
	switch cmd.OptionName {
	case "type":
		t := dns.GetType(cmd.OptionValue)
		if t == 0 {
			fmt.Printf("invalid query type: %s\n", cmd.OptionValue)
			return
		}

		params.QueryType = []dns.Type{t}

	case "port":
		port, err := strconv.Atoi(cmd.OptionValue)
		if err != nil {
			fmt.Printf("invalid port '%s': not a valid number", cmd.OptionValue)
			return
		}

		if port < 0 || port >= 65536 {
			fmt.Printf("invalid port '%s': out of range", cmd.OptionValue)
			return
		}

		params.Port = port
		_ = params.ReloadResolver()

		// case "vc":
		// 	params.TCP = true
		// 	_ = params.ReloadResolver()

		// case "novc":
		// 	params.TCP = false
		// 	_ = params.ReloadResolver()
	}
}

func updateParams(params *Params, cmd Command) bool {
	switch cmd.Command {
	case "exit":
		return true

	case "server":
		params.Server = cmd.OptionName
		err := params.ReloadResolver()
		if err != nil {
			fmt.Printf("invalid server IP: %s\n", err)
			return false
		}

		fmt.Printf("server:\t\t%s\n", params.Resolver.Endpoints[0])
		fmt.Printf("Address:\t%s\n", params.Resolver.Endpoints[0].FullAddress())

	case "set":
		updateParamsSet(params, cmd)
	}

	return false
}

func MainClassicalInteractiveLoop(finished chan<- struct{}) {
	params := NewDefaultParams()
	err := params.ReloadResolver()
	if err != nil {
		fmt.Printf("failed to load resolver: %v\n", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

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
			exit := updateParams(params, cmd)
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
