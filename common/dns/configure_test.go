package dns

import (
	"testing"

	"net"
	"strings"
)

func TestParseResolveConfContentBasic(t *testing.T) {
	content := strings.Join([]string{
		`# lorem ipsum`,
		`    `,
		"nameserver 1.2.3.4",
	}, "\n")

	conf, err := ParseResolvConfContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Nameservers) != 1 {
		t.Fatalf("expected 1 nameserver, got %d", len(conf.Nameservers))
	}

	expected := net.ParseIP("1.2.3.4")
	if !conf.Nameservers[0].Equal(expected) {
		t.Fatalf("expected nameserver %s, got %v", expected, conf.Nameservers[0])
	}
}

func TestParseResolveConfContentWithMultipleNameservers(t *testing.T) {
	content := strings.Join([]string{
		`nameserver 1.1.1.1`,
		`nameserver 2.2.2.2`,
	}, "\n")

	conf, err := ParseResolvConfContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Nameservers) != 2 {
		t.Fatalf("expected 2 nameservers, got %d", len(conf.Nameservers))
	}

	expected1 := net.ParseIP("1.1.1.1")
	expected2 := net.ParseIP("2.2.2.2")
	if !conf.Nameservers[0].Equal(expected1) {
		t.Fatalf("expected first nameserver %s, got %v", expected1, conf.Nameservers[0])
	}

	if !conf.Nameservers[1].Equal(expected2) {
		t.Fatalf("expected second nameserver %s, got %v", expected2, conf.Nameservers[1])
	}
}

func TestParseResolveConfContentInvalidNameserver(t *testing.T) {
	content := strings.Join([]string{
		`nameserver invalid_ip`,
	}, "\n")

	_, err := ParseResolvConfContent(content)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErrMsg := "invalid nameserver address format: nameserver invalid_ip"
	if err.Error() != expectedErrMsg {
		t.Fatalf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
