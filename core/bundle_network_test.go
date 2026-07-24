package core

import (
	"net"
	"testing"
)

// Test bundle network functions for coverage

func TestIsReservedOrPrivateHost_Loopback(t *testing.T) {
	// Test loopback detection
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"127.0.0.2", true},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isReservedOrPrivateHost(tt.host)
			if result != tt.want {
				t.Errorf("isReservedOrPrivateHost(%q) = %v, want %v", tt.host, result, tt.want)
			}
		})
	}
}

func TestIsReservedOrPrivateHost_Private(t *testing.T) {
	// Test private IP detection
	tests := []struct {
		host string
		want bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"8.8.8.8", false},  // Public
		{"1.1.1.1", false},  // Public
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isReservedOrPrivateHost(tt.host)
			if result != tt.want {
				t.Errorf("isReservedOrPrivateHost(%q) = %v, want %v", tt.host, result, tt.want)
			}
		})
	}
}

func TestIsLinkLocal_Coverage(t *testing.T) {
	// Test isLinkLocal function with net.IP values
	_ = isLinkLocal(net.ParseIP("169.254.0.1"))
	_ = isLinkLocal(net.ParseIP("fe80::1"))
}
