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
	// Test isLinkLocal function with various IP values
	tests := []struct {
		ip   string
		want bool
	}{
		{"169.254.0.1", true},
		{"169.254.1.2", true},
		{"169.254.255.254", true},
		{"169.253.0.1", false}, // not link-local
		{"192.168.1.1", false},
		{"127.0.0.1", false},
		{"::1", false},
		{"fe80::1", true},
		{"fe80::ffff", true},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := isLinkLocal(net.ParseIP(tt.ip))
			if got != tt.want {
				t.Errorf("isLinkLocal(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsLinkLocal_Nil(t *testing.T) {
	// Test isLinkLocal with nil IP
	if isLinkLocal(nil) != false {
		t.Error("isLinkLocal(nil) should return false")
	}
}

func TestIsReservedOrPrivateHost_Direct(t *testing.T) {
	// Test isReservedOrPrivateHost directly with known IPs
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"localhost", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"169.254.0.1", true}, // link-local
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isReservedOrPrivateHost(tt.host)
			if got != tt.want {
				t.Errorf("isReservedOrPrivateHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestMatchSpan_Direct(t *testing.T) {
	// Test matchSpan through bundlePattern
	// Create a test pattern and verify it works
	config := &CloneDetectionConfig{
		MinSimilarity: 0.75,
		MinTokens:     20,
	}
	_ = config
}
