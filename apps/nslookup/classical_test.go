package nslookup

import (
	"testing"
)

func checkCommand(t *testing.T, input string, expected Command) {
	t.Helper()

	cmd := ParseCommand(input)
	if cmd != expected {
		t.Errorf("Got wrong answer on ParseCommand(): %s", input)
		t.Errorf("expected: %+v", expected)
		t.Errorf("got:      %+v", cmd)
	}
}

func TestParseCommandWithSimpleDomain(t *testing.T) {
	input := "www.example.com"
	expected := Command{
		Command:     "",
		OptionName:  "",
		OptionValue: "",
		Name:        "www.example.com",
	}

	checkCommand(t, input, expected)
}

func TestParseCommandWithSimpleCommand(t *testing.T) {
	input := "help"
	expected := Command{
		Command:     "help",
		OptionName:  "",
		OptionValue: "",
		Name:        "",
	}

	checkCommand(t, input, expected)
}

func TestParseCommandWithOptionServer(t *testing.T) {
	input := "server 1.2.3.4"
	expected := Command{
		Command:     "server",
		OptionName:  "1.2.3.4",
		OptionValue: "",
		Name:        "",
	}

	checkCommand(t, input, expected)
}

func TestParseCommandWithOptionSetDebug(t *testing.T) {
	input := "set debug"
	expected := Command{
		Command:     "set",
		OptionName:  "debug",
		OptionValue: "",
		Name:        "",
	}

	checkCommand(t, input, expected)
}

func TestParseCommandWithOptionSetType(t *testing.T) {
	input := "set type=AAAA"
	expected := Command{
		Command:     "set",
		OptionName:  "type",
		OptionValue: "AAAA",
		Name:        "",
	}

	checkCommand(t, input, expected)
}
