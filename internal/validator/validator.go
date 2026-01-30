package validator

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}

	if len(rawURL) > 2048 {
		return fmt.Errorf("URL too long (max 2048 characters)")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	// Check host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	// SSRF protection
	if err := checkSSRF(parsed.Hostname()); err != nil {
		return err
	}

	return nil
}

func checkSSRF(hostname string) error {
	if os.Getenv("ALLOW_PRIVATE_IPS") == "true" {
		return nil
	}
	// Resolve hostname
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("could not resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("access to private IP addresses is not allowed")
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	// Check for private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // link-local
		"::1/128",        // IPv6 localhost
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
