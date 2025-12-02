# DenyIP Plugin for Traefik

Middleware to deny requests based on IP address. Supports IPv4 and IPv6 addresses in single IP and CIDR notation.

## Installation
```yaml
experimental:
  plugins:
    denyip:
      moduleName: github.com/devops9838/traefik-plugin-denyip
      version: v1.0.0
```

## Configuration
```yaml
http:
  middlewares:
    denyip:
      plugin:
        denyip:
          ipDenyList:
            - "192.168.1.0/24"
            - "10.0.0.1"
```

## Usage

Add to your router in traefik configuration.
