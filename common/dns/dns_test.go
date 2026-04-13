package dns

import (
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
