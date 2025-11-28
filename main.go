package traefik_plugin_blockip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Config holds the plugin configuration
type Config struct {
	BlockedIPs     []string `json:"blockedIPs,omitempty"`
	BlockedCIDRs   []string `json:"blockedCIDRs,omitempty"`
	WhitelistIPs   []string `json:"whitelistIPs,omitempty"`
	WhitelistCIDRs []string `json:"whitelistCIDRs,omitempty"`
	StatusCode     int      `json:"statusCode,omitempty"`
	Message        string   `json:"message,omitempty"`
	Debug          bool     `json:"debug,omitempty"`
	CacheTTL       int      `json:"cacheTTL,omitempty"`
}

// CreateConfig creates the default plugin configuration
func CreateConfig() *Config {
	return &Config{
		BlockedIPs:     []string{},
		BlockedCIDRs:   []string{},
		WhitelistIPs:   []string{},
		WhitelistCIDRs: []string{},
		StatusCode:     403,
		Message:        "Access Denied",
		Debug:          false,
		CacheTTL:       300,
	}
}

// IPCache provides fast caching for IP lookup results
type IPCache struct {
	mu    sync.RWMutex
	cache map[string]CacheEntry
}

// CacheEntry represents a cached lookup result
type CacheEntry struct {
	Status    string // "allowed", "blocked", "whitelisted"
	Timestamp int64
}

// BlockIP is the main plugin handler
type BlockIP struct {
	next   http.Handler
	name   string
	config *Config
	cache  *IPCache
}

// New creates a new BlockIP plugin instance
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &BlockIP{
		next:   next,
		name:   name,
		config: config,
		cache: &IPCache{
			cache: make(map[string]CacheEntry),
		},
	}, nil
}

// ServeHTTP implements the http.Handler interface
func (b *BlockIP) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	clientIP := b.getClientIP(req)
	
	if b.config.Debug {
		fmt.Printf("[BlockIP] Processing request from IP: %s\n", clientIP)
	}
	
	// Check whitelist first (highest priority)
	if b.isWhitelisted(clientIP) {
		if b.config.Debug {
			fmt.Printf("[BlockIP] IP %s is whitelisted, allowing\n", clientIP)
		}
		b.next.ServeHTTP(rw, req)
		return
	}
	
	// Check blocked list
	if b.isBlocked(clientIP) {
		if b.config.Debug {
			fmt.Printf("[BlockIP] IP %s is blocked, rejecting\n", clientIP)
		}
		rw.WriteHeader(b.config.StatusCode)
		rw.Write([]byte(b.config.Message))
		return
	}
	
	// Not blocked, allow
	b.next.ServeHTTP(rw, req)
}

// getClientIP extracts the client IP from the request
func (b *BlockIP) getClientIP(req *http.Request) string {
	// Check X-Forwarded-For first
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	
	// Check CF-Connecting-IP (Cloudflare)
	if cfIP := req.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}
	
	// Fall back to RemoteAddr
	if ra := req.RemoteAddr; ra != "" {
		host, _, err := net.SplitHostPort(ra)
		if err == nil {
			return host
		}
		return ra
	}
	
	return ""
}

// isWhitelisted checks if IP is in whitelist
func (b *BlockIP) isWhitelisted(ip string) bool {
	// Check direct IP whitelist
	for _, whiteIP := range b.config.WhitelistIPs {
		if ip == whiteIP {
			return true
		}
	}
	
	// Check whitelist CIDR ranges
	for _, cidr := range b.config.WhitelistCIDRs {
		if b.matchCIDR(ip, cidr) {
			return true
		}
	}
	
	return false
}

// isBlocked checks if IP is blocked
func (b *BlockIP) isBlocked(ip string) bool {
	// Check direct IP block list
	for _, blockedIP := range b.config.BlockedIPs {
		if ip == blockedIP {
			return true
		}
	}
	
	// Check blocked CIDR ranges
	for _, cidr := range b.config.BlockedCIDRs {
		if b.matchCIDR(ip, cidr) {
			return true
		}
	}
	
	return false
}

// matchCIDR checks if IP matches CIDR range
func (b *BlockIP) matchCIDR(ip string, cidr string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		if b.config.Debug {
			fmt.Printf("[BlockIP] Invalid CIDR %s: %v\n", cidr, err)
		}
		return false
	}
	
	return ipnet.Contains(parsedIP)
}
