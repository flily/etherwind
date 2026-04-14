package dns

import (
	"slices"
	"testing"
)

func TestCanonicalizeName(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{"example.com", "example.com."},
		{"example.com.", "example.com."},
		{"", "."},
	}

	for _, c := range cases {
		got := CanonicalizeName(c.name)
		if got != c.expected {
			t.Errorf("CanonicalizeName(%q) = %q, expected %q", c.name, got, c.expected)
		}
	}
}

func TestGetType(t *testing.T) {
	cases := []struct {
		name     string
		expected Type
	}{
		{"A", TypeA},
		{"AAAA", TypeAAAA},
		{"CNAME", TypeCNAME},
		{"MX", TypeMX},
		{"NS", TypeNS},
		{"UNKNOWN", Type(0)},
	}

	for _, c := range cases {
		got := GetType(c.name)
		if got != c.expected {
			t.Errorf("GetType(%q) = %d, expected %d", c.name, got, c.expected)
		}
	}
}

func TestParseTypes(t *testing.T) {
	cases := []struct {
		name     string
		expected []Type
	}{
		{"A", []Type{TypeA}},
		{"AAAA", []Type{TypeAAAA}},
		{"A+AAAA", []Type{TypeA, TypeAAAA}},
	}

	for _, c := range cases {
		got := ParseTypes(c.name)
		if len(got) != len(c.expected) {
			t.Errorf("ParseTypes(%q) = %v, expected %v", c.name, got, c.expected)
			continue
		}

		if !slices.Equal(got, c.expected) {
			t.Errorf("ParseTypes(%q) = %v, expected %v", c.name, got, c.expected)
		}
	}
}
