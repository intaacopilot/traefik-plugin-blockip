package traefik_plugin_blockip

import (
	"net"
	"strings"
)

// IPUtils provides utility functions for IP operations
type IPUtils struct{}

// ValidateIP validates if a string is a valid IP address
func (u *IPUtils) ValidateIP(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false
	}
	return net.ParseIP(ip) != nil
}

// ValidateCIDR validates if a string is a valid CIDR range
func (u *IPUtils) ValidateCIDR(cidr string) bool {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return false
	}
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IsIPv4 checks if an IP is IPv4
func (u *IPUtils) IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() != nil
}

// IsIPv6 checks if an IP is IPv6
func (u *IPUtils) IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() == nil && parsedIP.To16() != nil
}

// ExtractIPFromString extracts the first valid IP from a comma-separated string
func (u *IPUtils) ExtractIPFromString(s string) string {
	ips := strings.Split(s, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if u.ValidateIP(ip) {
			return ip
		}
	}
	return ""
}