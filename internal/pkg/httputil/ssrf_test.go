package httputil

import (
	"net"
	"testing"
)

func TestIsPublicIPAddress(t *testing.T) {
	test := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "Localhost 127.0.0.1",
			address:  "127.0.0.1",
			expected: false,
		},
		{
			name:     "AWS metadata endpoint 169.254.169.254",
			address:  "169.254.169.254",
			expected: false,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			ipaddress := net.ParseIP(tt.address)
			if ipaddress == nil {
				t.Fatalf("failed to parse IP address")
			}

			isPublic := IsPublicIPAddress(ipaddress)
			if isPublic != tt.expected {
				t.Errorf("IsPublicIPAddress(%v) = %v, want %v", ipaddress, isPublic, tt.expected)
			}
		})
	}
}
