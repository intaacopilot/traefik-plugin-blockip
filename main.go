package traefik_plugin_blockip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
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

// ipLookupService encapsulates IP lookup logic
type ipLookupService struct {
	blockedIPsSet    map[string]bool
	blockedNets      []*net.IPNet
	whitelistIPsSet  map[string]bool
	whitelistNets    []*net.IPNet
	cache            *IPCache
	cacheTTL         int64
	mu               sync.RWMutex
}

// BlockIP is the main plugin handler
type BlockIP struct {
	next          http.Handler
	name          string
	lookup        *ipLookupService
	statusCode    int
	message       string
	debug         bool
	responseBody  []byte
}

// New creates and returns a new BlockIP plugin instance
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if next == nil {
		return nil, fmt.Errorf("next handler is nil")
	}

	plugin := &BlockIP{
		next:   next,
		name:   name,
		debug:  config.Debug,
		lookup: newIPLookupService(config.CacheTTL),
	}

	// Set status code
	if config.StatusCode == 0 {
		plugin.statusCode = 403
	} else if config.StatusCode < 400 || config.StatusCode >= 600 {
		return nil, fmt.Errorf("invalid status code: %d, must be 4xx or 5xx", config.StatusCode)
	} else {
		plugin.statusCode = config.StatusCode
	}

	// Set message
	if config.Message == "" {
		plugin.message = "Access Denied"
	} else {
		plugin.message = config.Message
	}

	plugin.responseBody = []byte(plugin.message)

	// Parse and load IP configurations
	if err := plugin.loadConfiguration(config); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if plugin.debug {
		fmt.Printf("[%s] Plugin initialized with status code %d\n", plugin.name, plugin.statusCode)
	}

	return plugin, nil
}

// newIPLookupService creates a new IP lookup service
func newIPLookupService(cacheTTL int) *ipLookupService {
	if cacheTTL <= 0 {
		cacheTTL = 300
	}

	return &ipLookupService{
		blockedIPsSet:   make(map[string]bool),
		blockedNets:     make([]*net.IPNet, 0),
		whitelistIPsSet: make(map[string]bool),
		whitelistNets:   make([]*net.IPNet, 0),
		cache: &IPCache{
			cache: make(map[string]CacheEntry),
		},
		cacheTTL: int64(cacheTTL),
	}
}

// loadConfiguration parses and loads the configuration
func (p *BlockIP) loadConfiguration(config *Config) error {
	// Parse blocked IPs
	for _, ip := range config.BlockedIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if ! isValidIP(ip) {
			if p.debug {
				fmt.Printf("[%s] Invalid IP format: %s\n", p.name, ip)
			}
			continue
		}

		p.lookup.blockedIPsSet[ip] = true
		if p.debug {
			fmt.Printf("[%s] Added blocked IP: %s\n", p.name, ip)
		}
	}

	// Parse blocked CIDRs
	for _, cidr := range config.BlockedCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		if err := p.parseCIDR(cidr, true); err != nil {
			if p.debug {
				fmt.Printf("[%s] Error parsing blocked CIDR %s: %v\n", p.name, cidr, err)
			}
			continue
		}

		if p.debug {
			fmt.Printf("[%s] Added blocked CIDR: %s\n", p.name, cidr)
		}
	}

	// Parse whitelist IPs
	for _, ip := range config.WhitelistIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if !isValidIP(ip) {
			if p.debug {
				fmt.Printf("[%s] Invalid whitelist IP format: %s\n", p.name, ip)
			}
			continue
		}

		p.lookup.whitelistIPsSet[ip] = true
		if p.debug {
			fmt.Printf("[%s] Added whitelist IP: %s\n", p.name, ip)
		}
	}

	// Parse whitelist CIDRs
	for _, cidr := range config.WhitelistCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		if err := p.parseCIDR(cidr, false); err != nil {
			if p.debug {
				fmt.Printf("[%s] Error parsing whitelist CIDR %s: %v\n", p.name, cidr, err)
			}
			continue
		}

		if p.debug {
			fmt.Printf("[%s] Added whitelist CIDR: %s\n", p.name, cidr)
		}
	}

	if p.debug {
		fmt.Printf("[%s] Configuration loaded.Blocked IPs: %d, Blocked CIDRs: %d, Whitelist IPs: %d, Whitelist CIDRs: %d\n",
			p.name,
			len(p.lookup.blockedIPsSet),
			len(p.lookup.blockedNets),
			len(p.lookup.whitelistIPsSet),
			len(p.lookup.whitelistNets),
		)
	}

	return nil
}

// parseCIDR parses a CIDR range and adds it to the appropriate list
func (p *BlockIP) parseCIDR(cidr string, isBlocked bool) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %w", err)
	}

	if isBlocked {
		p.lookup.blockedNets = append(p.lookup.blockedNets, ipnet)
	} else {
		p.lookup.whitelistNets = append(p.lookup.whitelistNets, ipnet)
	}

	return nil
}

// ServeHTTP implements the http.Handler interface
func (p *BlockIP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := p.getClientIP(r)

	if p.debug {
		fmt.Printf("[%s] Request from IP: %s, Path: %s\n", p.name, clientIP, r.RequestURI)
	}

	if clientIP == "" {
		if p.debug {
			fmt.Printf("[%s] Could not extract client IP\n", p.name)
		}
		p.next.ServeHTTP(w, r)
		return
	}

	// Check cache first for faster response
	if status := p.lookup.checkCache(clientIP); status != "" {
		if p.debug {
			fmt.Printf("[%s] Cache hit for IP %s: %s\n", p.name, clientIP, status)
		}

		switch status {
		case "whitelisted":
			p.next.ServeHTTP(w, r)
			return
		case "blocked":
			p.sendBlockResponse(w)
			return
		case "allowed":
			p.next.ServeHTTP(w, r)
			return
		}
	}

	// Check if IP is whitelisted (priority 1)
	if p.lookup.isWhitelisted(clientIP) {
		if p.debug {
			fmt.Printf("[%s] IP %s is whitelisted\n", p.name, clientIP)
		}
		p.lookup.cacheResult(clientIP, "whitelisted")
		p.next.ServeHTTP(w, r)
		return
	}

	// Check if IP is blocked (priority 2)
	if p.lookup.isBlocked(clientIP) {
		if p.debug {
			fmt.Printf("[%s] IP %s is blocked\n", p.name, clientIP)
		}
		p.lookup.cacheResult(clientIP, "blocked")
		p.sendBlockResponse(w)
		return
	}

	// Allowed by default
	if p.debug {
		fmt.Printf("[%s] IP %s is allowed (not blocked)\n", p.name, clientIP)
	}
	p.lookup.cacheResult(clientIP, "allowed")
	p.next.ServeHTTP(w, r)
}

// sendBlockResponse sends a block response to the client
func (p *BlockIP) sendBlockResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(p.responseBody)))
	w.WriteHeader(p.statusCode)
	_, err := w.Write(p.responseBody)
	if err != nil && p.debug {
		fmt.Printf("[%s] Error writing response: %v\n", p.name, err)
	}
}

// getClientIP extracts the client IP from the request with proper error handling
func (p *BlockIP) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (common with reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if isValidIP(ip) {
				if p.debug {
					fmt.Printf("[%s] Extracted IP from X-Forwarded-For: %s\n", p.name, ip)
				}
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		xri = strings.TrimSpace(xri)
		if isValidIP(xri) {
			if p.debug {
				fmt.Printf("[%s] Extracted IP from X-Real-IP: %s\n", p.name, xri)
			}
			return xri
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		cfip = strings.TrimSpace(cfip)
		if isValidIP(cfip) {
			if p.debug {
				fmt.Printf("[%s] Extracted IP from CF-Connecting-IP: %s\n", p.name, cfip)
			}
			return cfip
		}
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if p.debug {
		fmt.Printf("[%s] Using RemoteAddr: %s\n", p.name, ip)
	}

	if strings.Contains(ip, ":") {
		var err error
		ip, _, err = net.SplitHostPort(ip)
		if err != nil {
			if p.debug {
				fmt.Printf("[%s] Error parsing RemoteAddr %s: %v\n", p.name, r.RemoteAddr, err)
			}
			return ""
		}
	}

	if ! isValidIP(ip) {
		if p.debug {
			fmt.Printf("[%s] Invalid IP extracted: %s\n", p.name, ip)
		}
		return ""
	}

	if p.debug {
		fmt.Printf("[%s] Extracted IP from RemoteAddr: %s\n", p.name, ip)
	}

	return ip
}

// isValidIP checks if a string is a valid IP address
func isValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	return net.ParseIP(ip) != nil
}

// checkCache retrieves cached lookup result
func (s *ipLookupService) checkCache(ip string) string {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	entry, exists := s.cache.cache[ip]
	if !exists {
		return ""
	}

	// Check if cache entry is still valid
	now := time.Now().Unix()
	if now-entry.Timestamp > s.cacheTTL {
		return ""
	}

	return entry.Status
}

// cacheResult stores a lookup result in cache
func (s *ipLookupService) cacheResult(ip string, status string) {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	s.cache.cache[ip] = CacheEntry{
		Status:    status,
		Timestamp: time.Now().Unix(),
	}

	// Cleanup old entries if cache gets too large
	if len(s.cache.cache) > 10000 {
		s.cleanupCache()
	}
}

// cleanupCache removes old entries from cache (must be called with lock held)
func (s *ipLookupService) cleanupCache() {
	now := time.Now().Unix()
	count := 0
	for ip, entry := range s.cache.cache {
		if now-entry.Timestamp > s.cacheTTL {
			delete(s.cache.cache, ip)
			count++
		}
	}
}

// isWhitelisted checks if the IP is whitelisted with optimized lookup
func (s *ipLookupService) isWhitelisted(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check direct IP match first (O(1) operation)
	if s.whitelistIPsSet[ip] {
		return true
	}

	// Check CIDR ranges (O(n) operation)
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, ipnet := range s.whitelistNets {
		if ipnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// isBlocked checks if the IP is blocked with optimized lookup
func (s *ipLookupService) isBlocked(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check direct IP match first (O(1) operation)
	if s.blockedIPsSet[ip] {
		return true
	}

	// Check CIDR ranges (O(n) operation)
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, ipnet := range s.blockedNets {
		if ipnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}