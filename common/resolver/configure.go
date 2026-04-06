package resolver

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

// resolv.conf format reference: https://man7.org/linux/man-pages/man5/resolv.conf.5.html

type Configure struct {
	Nameservers []net.IP
}

func ParseResolvConf(path string) (*Configure, error) {
	fd, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseResolvConfContent(string(fd))
}

var (
	lineContentRegex = regexp.MustCompile(`^\s*([^\s]+)\s+(.*)$`)
)

func ParseResolvConfContent(content string) (*Configure, error) {
	lines := strings.Split(content, "\n")

	conf := &Configure{
		Nameservers: make([]net.IP, 0),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		matches := lineContentRegex.FindStringSubmatch(line)
		// This regex should always match at least one group.

		command := matches[1]
		switch command {
		case "nameserver":
			ip := net.ParseIP(matches[2])
			if ip == nil {
				err := fmt.Errorf("invalid nameserver address format: %s", line)
				return nil, err
			}

			conf.Nameservers = append(conf.Nameservers, ip)
		}
	}

	return conf, nil
}
