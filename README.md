# Traefik BlockIP Plugin v2.0

A high-performance, production-ready Traefik middleware plugin for blocking or allowing HTTP requests based on client IP addresses with integrated caching and comprehensive error handling.

## Features

- ✅ **Block individual IP addresses and CIDR ranges** (IPv4 & IPv6)
- ✅ **Whitelist IP addresses and ranges** with priority checking
- ✅ **Request caching** for ultra-fast lookups (O(1) performance)
- ✅ **Multiple header support** (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
- ✅ **Comprehensive error handling** with custom error codes
- ✅ **Debug logging** with log buffering
- ✅ **Thread-safe operations** with mutex-based synchronization
- ✅ **Optimized lookup** with hash maps and CIDR matching
- ✅ **Configurable cache TTL** for memory management
- ✅ **Production-ready** with extensive validation

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────┐
│           HTTP Request                          │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
         ┌─────────────────────┐
         │  getClientIP()      │
         │  (Extract from      │
         │   headers/RemoteAddr)
         └──────────┬──────────┘
                    │
                    ▼
         ┌──────────────────────┐
         │  checkCache()        │
         │  (Fast lookup O(1))  │
         └──────────┬───────────┘
                    │
         ┌──────────▼───────────┐
         │   Cache Hit?          │
         └──────────┬───────────┘
              ┌─────┴─────┐
           Yes│           │No
              ▼           ▼
        ┌─────────┐  ┌─────────────────────┐
        │ Return  │  │ isWhitelisted()     │
        │ Cached  │  │ (Check direct + net)
        │ Result  │  └──────────┬──────────┘
        └─────────┘             │
                      ┌─────────▼────────┐
                      │ Whitelist Match? │
                      └────────┬─────────┘
                            No │
                               ▼
                        ┌──────────────────┐
                        │ isBlocked()      │
                        │ (Check direct +  │
                        │  CIDR ranges)    │
                        └────────┬─────────┘
                                 │
                      ┌──────────▼────────┐
                      │ Block Match?      │
                      └────────┬─────────┘
                            Yes│ No
                               │
                    ┌──────────▼────────┐
                    │ cacheResult() &   │
                    │ Route Request     │
                    └───────────────────┘
```

## Installation

### Using Traefik Pilot

```bash
# Enable the plugin in your Traefik configuration
# See configuration examples below
```

### Manual Installation

```bash
# Clone repository
git clone https://github.com/intaacopilot/traefik-plugin-blockip. git
cd traefik-plugin-blockip

# Build
go build -o plugin. so .

# Test
go test -v ./...
```

## Configuration

### Basic Configuration

```yaml
middlewares:
  blockip:
    plugin:
      blockip:
        # Block specific IPs
        blockedIPs:
          - "192. 168.1.100"
          - "203.0.113.50"
        
        # Block CIDR ranges
        blockedCIDRs:
          - "192.168.0.0/16"
          - "10. 0.0.0/8"
        
        # HTTP response settings
        statusCode: 403
        message: "Access Denied"
        
        # Performance settings
        cacheTTL: 300
        
        # Debug mode
        debug: false
```

### Advanced Configuration

```yaml
middlewares:
  blockip-advanced:
    plugin:
      blockip:
        # Blocked IPs and ranges
        blockedIPs:
          - "192.168.1.100"
          - "192.168.1.101"
          - "203.0.113.50"
        
        blockedCIDRs:
          - "192.168.0. 0/16"    # Entire subnet
          - "10.0. 0.0/8"        # Large range
          - "2001:db8::/32"     # IPv6 range
        
        # Whitelist (bypass blocking)
        whitelistIPs:
          - "127. 0.0.1"
          - "::1"
          - "10.0.0. 1"
        
        whitelistCIDRs:
          - "192.168.1.0/24"    # Trusted subnet
          - "10.0. 0.0/16"       # Trusted network
          - "2001:db8:1::/48"   # IPv6 trusted range
        
        # Response configuration
        statusCode: 403
        message: "Your IP has been blocked"
        
        # Cache configuration (in seconds)
        cacheTTL: 300           # 5 minute cache
        
        # Debug logging
        debug: true
```

### Configuration Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `blockedIPs` | []string | No | `[]` | Individual IPs to block |
| `blockedCIDRs` | []string | No | `[]` | CIDR ranges to block (IPv4 & IPv6) |
| `whitelistIPs` | []string | No | `[]` | IPs to whitelist (bypass blocking) |
| `whitelistCIDRs` | []string | No | `[]` | CIDR ranges to whitelist |
| `statusCode` | int | No | `403` | HTTP status code (400-599) |
| `message` | string | No | `"Access Denied"` | Response message |
| `cacheTTL` | int | No | `300` | Cache duration in seconds |
| `debug` | bool | No | `false` | Enable debug logging |

## Usage Examples

### Docker Compose

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints. web.address=:80"
      - "--experimental.plugins.blockip.modulename=github.com/intaacopilot/traefik-plugin-blockip"
      - "--experimental.plugins. blockip.version=2.0.0"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  web-app:
    image: nginx:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers. web.rule=Host(`example. com`)"
      - "traefik.http.routers. web.entrypoints=web"
      - "traefik.http. middlewares.blockip.plugin.blockip.blockedCIDRs=192.168.0.0/16,10.0.0. 0/8"
      - "traefik.http. middlewares.blockip.plugin. blockip.whitelistIPs=192.168.1.1"
      - "traefik. http.middlewares.blockip. plugin.blockip.statusCode=403"
      - "traefik.http.middlewares.blockip.plugin.blockip.cacheTTL=300"
      - "traefik.http. middlewares.blockip.plugin. blockip.debug=true"
      - "traefik.http. routers.web.middlewares=blockip"
```

### Static Configuration (YAML)

```yaml
# traefik.yml
entryPoints:
  web:
    address: :80

providers:
  file:
    filename: /etc/traefik/config.yml
  docker: {}

experimental:
  plugins:
    blockip:
      modulename: github.com/intaacopilot/traefik-plugin-blockip
      version: 2.0.0
```

```yaml
# config.yml
http:
  middlewares:
    blockip-prod:
      plugin:
        blockip:
          blockedCIDRs:
            - "192.168.0.0/16"
            - "10. 0.0.0/8"
          whitelistCIDRs:
            - "192.168.1.0/24"
          statusCode: 403
          message: "IP Blocked"
          cacheTTL: 300
          debug: false
  
  routers:
    my-app:
      rule: "Host(`app.example.com`)"
      middlewares:
        - blockip-prod
      service: my-service
  
  services:
    my-service:
      loadBalancer:
        servers:
          - url: http://localhost:8080
```

### Kubernetes

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: blockip
spec:
  plugin:
    blockip:
      blockedCIDRs:
        - "192.168.0.0/16"
        - "10.0.0.0/8"
      whitelistIPs:
        - "10.0.0.1"
      whitelistCIDRs:
        - "192.168.1.0/24"
      statusCode: 403
      message: "Access Denied"
      cacheTTL: 300
      debug: false
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    traefik.ingress.kubernetes.io/router.middlewares: default-blockip@kubernetescrd
spec:
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 80
```

## Performance

### Lookup Performance

- **Direct IP Match**: O(1) - Hash map lookup
- **CIDR Range Match**: O(n) - Linear search through CIDR list
- **Cache Hit**: O(1) - Hash map lookup with TTL validation
- **Overall**: Sub-millisecond response time for most requests

### Memory Optimization

- **Request Cache**: Limited to 10,000 entries with automatic cleanup
- **Cache TTL**: Configurable (default 300 seconds)
- **Automatic Rotation**: Old entries removed when limit reached

## Error Handling

### Error Codes

| Code | Description | Resolution |
|------|-------------|-----------|
| `INVALID_CONFIG` | Configuration is missing or invalid | Check config structure |
| `NIL_HANDLER` | Next handler is nil | Verify handler chain setup |
| `INVALID_STATUS_CODE` | Status code outside 4xx-5xx range | Use 4xx or 5xx codes |
| `INVALID_CIDR` | CIDR format is incorrect | Validate CIDR syntax |
| `INVALID_IP` | IP format is incorrect | Validate IP format |
| `PARSE_ERROR` | Error parsing configuration | Check config syntax |
| `INTERNAL_ERROR` | Internal plugin error | Check debug logs |

### Exception Handling

```go
// Invalid configuration is caught at startup
if config.StatusCode < 400 || config. StatusCode >= 600 {
    return nil, fmt.Errorf("invalid status code: %d", config.StatusCode)
}

// Invalid CIDR is logged but doesn't crash plugin
if err := parseCIDR(cidr); err != nil {
    logger. Warn("Invalid CIDR: %s", cidr)
    continue
}

// IP parsing errors are handled gracefully
if ! isValidIP(ip) {
    logger.Debug("Invalid IP extracted: %s", ip)
    return ""
}
```

## Debug Mode

Enable debug logging for troubleshooting:

```yaml
middlewares:
  blockip-debug:
    plugin:
      blockip:
        debug: true
        # ... other config
```

Output example:
```
[2025-11-27 10:30:45] DEBUG - [blockip] Configuration loaded successfully.  Blocked IPs: 2, Blocked CIDRs: 2, Whitelist IPs: 1, Whitelist CIDRs: 1
[2025-11-27 10:30:46] DEBUG - [blockip] Incoming request from IP: 192.168.1. 100, Path: /api/users
[2025-11-27 10:30:46] DEBUG - [blockip] IP 192.168.1.100 is blocked
```

## Security Considerations

1. **Trust Headers**: Ensure X-Forwarded-For comes from trusted proxies
2. **Cache Poisoning**: Adjust cache TTL based on your security requirements
3. **Whitelist Precedence**: Whitelist is checked before block list
4. **Logging**: Enable debug mode in staging, disable in production to avoid logs
5. **Regular Updates**: Keep plugin updated for security patches

## Testing

```bash
# Run tests
go test -v ./... 

# Test with coverage
go test -v -cover ./...

# Build and test locally
go build -o plugin. so . 
```

## Troubleshooting

### IPs not being blocked

1. Verify IP format is correct (use debug mode)
2. Check if IP is whitelisted
3. Confirm CIDR notation is valid
4. Review logs with `debug: true`

### High latency

1.  Reduce number of CIDR ranges
2. Increase `cacheTTL` value
3. Monitor cache hit rate in debug logs

### Memory usage

1. Reduce `cacheTTL` or set to 0 to disable
2.  Reduce number of blocked/whitelisted IPs
3.  Check for cache leaks in debug logs

## Contributing

Contributions welcome! Please submit pull requests to improve the plugin.

## License

MIT License - See LICENSE file for details

## Support

For issues, features, or questions:
- GitHub Issues: https://github.com/intaacopilot/traefik-plugin-blockip/issues
- Documentation: https://github.com/intaacopilot/traefik-plugin-blockip

## Changelog

### v2.0.0 (2025-11-27)
- Added comprehensive error handling
- Implemented request caching for performance
- Added logger component with log buffering
- Created utils package for IP validation
- Thread-safe operations with mutex synchronization
- Support for Cloudflare IP header
- Improved documentation and examples

### v1.0.0 (2025-11-27)
- Initial release
- Basic IP blocking and whitelisting
- CIDR range support