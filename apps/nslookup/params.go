package nslookup

import (
	"time"
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
	Server  string
	TCP     bool
	Timeout time.Duration
}

func NewDefaultParams() *Params {
	p := &Params{
		Server:  "",
		TCP:     false,
		Timeout: 2 * time.Second,
	}

	return p
}
