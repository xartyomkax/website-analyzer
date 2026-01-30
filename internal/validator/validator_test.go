package validator

import (
	"net"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTPS", "https://www.google.com", false},
		{"Valid HTTP", "http://example.com/path?query=1", false},
		{"Empty URL", "", true},
		{"Invalid Scheme", "ftp://example.com", true},
		{"No Host", "https://", true},
		{"Too Long URL", "https://example.com/" + string(make([]byte, 2040)), true},
		{"Local IP 127.0.0.1", "http://127.0.0.1", true},
		{"Local IP 192.168.1.1", "http://192.168.1.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ipStr string
		want  bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.0.1", true},
		{"169.254.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
		{"fe80::1", true},
	}

	for _, tt := range tests {
		t.Run(tt.ipStr, func(t *testing.T) {
			ip := net.ParseIP(tt.ipStr)
			if got := isPrivateIP(ip); got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ipStr, got, tt.want)
			}
		})
	}
}
