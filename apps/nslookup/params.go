package nslookup

import (
	"time"
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
