package traefik_plugin_blockip

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.StatusCode = 403
	config.Message = "Blocked"

	handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	if handler == nil {
		t.Fatal("Handler is nil")
	}
}

func TestNewWithNilConfig(t *testing.T) {
	_, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil, "blockip-test")

	if err == nil {
		t.Fatal("Expected error for nil config")
	}
}

func TestNewWithNilHandler(t *testing.T) {
	config := CreateConfig()
	_, err := New(context.Background(), nil, config, "blockip-test")

	if err == nil {
		t.Fatal("Expected error for nil handler")
	}
}

func TestBlockIP(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestAllowIP(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.101:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCIDRBlocking(t *testing.T) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"192.168.0.0/16"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for CIDR blocking, got %d", w.Code)
	}
}

func TestWhitelistBypass(t *testing.T) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"192.168.0.0/16"}
	config.WhitelistIPs = []string{"192.168.1.50"}
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 (whitelisted), got %d", w.Code)
	}
}

func TestXForwardedForHeader(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"203.0.113.50"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for X-Forwarded-For header, got %d", w.Code)
	}
}

func TestInvalidStatusCode(t *testing.T) {
	config := CreateConfig()
	config.StatusCode = 200 // Invalid for error response

	_, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	if err == nil {
		t.Fatal("Expected error for invalid status code")
	}
}

func TestIPValidation(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
		testName string
	}{
		{"192.168.1.1", true, "Valid IPv4"},
		{"10.0.0.0", true, "Valid IPv4"},
		{"256.1.1.1", false, "Invalid IPv4"},
		{"::1", true, "Valid IPv6 loopback"},
		{"2001:db8::1", true, "Valid IPv6"},
		{"", false, "Empty string"},
		{"invalid", false, "Invalid string"},
	}

	for _, test := range tests {
		result := isValidIP(test.ip)
		if result != test.expected {
			t.Errorf("%s: expected %v, got %v", test.testName, test.expected, result)
		}
	}
}

func TestCIDRValidation(t *testing.T) {
	utils := &IPUtils{}

	tests := []struct {
		cidr     string
		expected bool
		testName string
	}{
		{"192.168.0.0/16", true, "Valid IPv4 CIDR"},
		{"10.0.0.0/8", true, "Valid IPv4 CIDR"},
		{"2001:db8::/32", true, "Valid IPv6 CIDR"},
		{"192.168.0.0/33", false, "Invalid CIDR (out of range)"},
		{"invalid/16", false, "Invalid CIDR format"},
		{"", false, "Empty string"},
	}

	for _, test := range tests {
		result := utils.ValidateCIDR(test.cidr)
		if result != test.expected {
			t.Errorf("%s: expected %v, got %v", test.testName, test.expected, result)
		}
	}
}

func TestMultipleXForwardedForIPs(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"203.0.113.50"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	// Multiple IPs, first one should be blocked
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 192.168.1.1, 10.0.0.2")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for blocked IP in X-Forwarded-For, got %d", w.Code)
	}
}

func TestXRealIPHeader(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"203.0.113.50"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.50")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for X-Real-IP header, got %d", w.Code)
	}
}

func TestWhitelistCIDR(t *testing.T) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"192.168.0.0/16"}
	config.WhitelistCIDRs = []string{"192.168.1.0/24"}
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 (whitelisted CIDR), got %d", w.Code)
	}
}

func TestIPv6Blocking(t *testing.T) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"2001:db8::/32"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for IPv6 CIDR, got %d", w.Code)
	}
}

func TestCFConnectingIPHeader(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"203.0.113.50"}
	config.StatusCode = 403
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("CF-Connecting-IP", "203.0.113.50")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("Expected status 403 for CF-Connecting-IP header, got %d", w.Code)
	}
}

func TestEmptyClientIP(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ""

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should allow when IP is empty
	if w.Code != 200 {
		t.Errorf("Expected status 200 for empty IP, got %d", w.Code)
	}
}

func TestWhitelistPriority(t *testing.T) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.WhitelistIPs = []string{"192.168.1.100"}
	config.Debug = false

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-test")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Whitelist should have priority
	if w.Code != 200 {
		t.Errorf("Expected status 200 (whitelist priority), got %d", w.Code)
	}
}

// Benchmarks
func BenchmarkIPLookupDirect(b *testing.B) {
	config := CreateConfig()
	config.BlockedIPs = make([]string, 100)
	for i := 0; i < 100; i++ {
		config.BlockedIPs[i] = fmt.Sprintf("192.168.%d.%d", i/256, i%256)
	}

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-bench")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.0.50:12345"

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkCIDRLookup(b *testing.B) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12"}

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-bench")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkCacheLookup(b *testing.B) {
	config := CreateConfig()
	config.BlockedIPs = []string{"192.168.1.100"}
	config.CacheTTL = 3600

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-bench")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()

	// Prime the cache
	handler.ServeHTTP(w, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkWhitelistCheck(b *testing.B) {
	config := CreateConfig()
	config.BlockedCIDRs = []string{"192.168.0.0/16"}
	config.WhitelistIPs = []string{"192.168.1.50"}

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-bench")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkHeaderExtraction(b *testing.B) {
	config := CreateConfig()
	config.BlockedIPs = []string{"203.0.113.50"}

	handler, _ := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "blockip-bench")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 192.168.1.1")

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}