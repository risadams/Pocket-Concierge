# Configuration Reference

This document provides a complete reference for configuring Pocket Concierge DNS server. The configuration is defined in YAML format and controls all aspects of server behavior.

## Configuration File Location

By default, Pocket Concierge looks for `config.yaml` in the current working directory. You can specify a custom configuration file path:

```bash
./pocketconcierge /path/to/your/config.yaml
```

## Complete Configuration Example

```yaml
# Server configuration
server:
  port: 8053                    # Port to listen on (default: 8053)
  address: "127.0.0.1"          # IP address to bind to (default: 127.0.0.1)

# DNS server settings
dns:
  ttl: 300                      # Default TTL for responses in seconds
  enable_recursion: true        # Allow recursive queries
  cache_size: 10000             # Maximum number of cached entries
  block_list:                   # Domains to block
    - "ads.example.com"
    - "tracker.badsite.org"
    - "malware.evil.net"

# Upstream DNS servers (in order of preference)
upstream:
  # Primary: DNS-over-HTTPS
  - name: "ControlD DoH Primary"
    address: "dns.controld.com"
    protocol: "https"
    port: 443
    path: "/YOUR_ENDPOINT_ID"   # Your ControlD resolver endpoint
    verify: true                # Verify TLS certificates

  # Secondary: DNS-over-TLS
  - name: "Cloudflare DoT"
    address: "1dot1dot1dot1.cloudflare-dns.com"
    protocol: "tls"
    port: 853
    verify: true

  # Fallback: Traditional DNS
  - name: "Cloudflare Primary"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53
    verify: false

  - name: "Cloudflare Secondary"
    address: "1.0.0.1"
    protocol: "udp"
    port: 53
    verify: false

# Logging configuration
log_level: "info"               # Log levels: debug, info, warn, error

# Home network configuration
home_dns_domain: "home"         # Default domain suffix for simple hostnames

# Local host mappings
hosts:
  # Simple hostname (will become "desktop.home")
  - hostname: "desktop"
    ipv4:
      - "192.168.1.100"
    ipv6:
      - "fe80::1234:5678:90ab:cdef"

  # Multiple IP addresses
  - hostname: "server"
    ipv4:
      - "192.168.1.10"
      - "192.168.1.11"           # Load balancing/redundancy

  # IPv4 only
  - hostname: "printer"
    ipv4:
      - "192.168.1.53"

  # Full domain name (won't get .home suffix)
  - hostname: "nas.local"
    ipv4:
      - "192.168.1.20"

  # IoT devices
  - hostname: "thermostat"
    ipv4:
      - "192.168.1.75"

  - hostname: "camera1"
    ipv4:
      - "192.168.1.80"

  - hostname: "camera2"
    ipv4:
      - "192.168.1.81"
```

## Configuration Sections

### Server Configuration

Controls the DNS server's network behavior.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `port` | integer | `8053` | UDP/TCP port to listen on for DNS queries |
| `address` | string | `"127.0.0.1"` | IP address to bind to. Use `"0.0.0.0"` for all interfaces |

**Examples:**

```yaml
server:
  port: 53                      # Standard DNS port (requires root/admin)
  address: "0.0.0.0"           # Listen on all network interfaces
```

### DNS Configuration

Controls DNS protocol behavior and caching.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `ttl` | integer | `300` | Default TTL (Time To Live) for DNS responses in seconds |
| `enable_recursion` | boolean | `true` | Enable recursive DNS queries |
| `cache_size` | integer | `10000` | Maximum number of entries in the DNS cache |
| `block_list` | array | `[]` | List of domains to block (return NXDOMAIN) |

**Examples:**

```yaml
dns:
  ttl: 600                      # 10 minutes TTL
  enable_recursion: false       # Disable recursion (authoritative only)
  cache_size: 50000            # Larger cache for busy networks
  block_list:                   # Block unwanted domains
    - "ads.example.com"         # Block specific subdomain
    - "tracker.badsite.org"     # Block tracking domains
    - "malware.net"             # Block entire domain and all subdomains
```

#### Block List Behavior

The `block_list` feature allows you to prevent resolution of unwanted domains:

- **Exact matches**: `"evil.com"` in the block list will block `evil.com`
- **Subdomain blocking**: `"evil.com"` will also block `sub.evil.com`, `deep.sub.evil.com`, etc.
- **Case insensitive**: Matching is performed in a case-insensitive manner
- **NXDOMAIN response**: Blocked domains return a DNS NXDOMAIN (Name Error) response
- **No upstream forwarding**: Blocked domains are not forwarded to upstream servers

**Use cases:**

- Ad blocking: Block advertising domains
- Malware protection: Block known malicious domains  
- Parental controls: Block inappropriate content domains
- Privacy protection: Block tracking and analytics domains

### Upstream DNS Servers

Configure external DNS servers for resolving non-local queries. Servers are tried in order until one responds successfully.

#### Common Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Descriptive name for the upstream server |
| `address` | string | Yes | Hostname or IP address of the DNS server |
| `protocol` | string | Yes | Protocol: `udp`, `tcp`, `tls`, or `https` |
| `port` | integer | Yes | Port number for the DNS server |
| `verify` | boolean | No | Verify TLS certificates (for `tls` and `https` protocols) |

#### Protocol-Specific Parameters

**DNS-over-HTTPS (`https`)**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | URL path for the DoH endpoint |

**DNS-over-TLS (`tls`)**:
No additional parameters required.

**Traditional DNS (`udp`/`tcp`)**:
No additional parameters required.

#### Popular Upstream Servers

**Cloudflare:**

```yaml
# DNS-over-HTTPS
- name: "Cloudflare DoH"
  address: "cloudflare-dns.com"
  protocol: "https"
  port: 443
  path: "/dns-query"
  verify: true

# DNS-over-TLS
- name: "Cloudflare DoT"
  address: "1dot1dot1dot1.cloudflare-dns.com"
  protocol: "tls"
  port: 853
  verify: true

# Traditional DNS
- name: "Cloudflare UDP"
  address: "1.1.1.1"
  protocol: "udp"
  port: 53
```

**Google Public DNS:**

```yaml
# DNS-over-HTTPS
- name: "Google DoH"
  address: "dns.google"
  protocol: "https"
  port: 443
  path: "/dns-query"
  verify: true

# Traditional DNS
- name: "Google UDP"
  address: "8.8.8.8"
  protocol: "udp"
  port: 53
```

**Quad9:**

```yaml
# DNS-over-HTTPS
- name: "Quad9 DoH"
  address: "dns.quad9.net"
  protocol: "https"
  port: 443
  path: "/dns-query"
  verify: true

# Traditional DNS
- name: "Quad9 UDP"
  address: "9.9.9.9"
  protocol: "udp"
  port: 53
```

### Logging Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `log_level` | string | `"info"` | Logging verbosity: `debug`, `info`, `warn`, `error` |

**Examples:**

```yaml
log_level: "debug"              # Verbose logging for troubleshooting
log_level: "error"              # Minimal logging for production
```

### Home Network Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `home_dns_domain` | string | `"home"` | Domain suffix automatically added to simple hostnames |

When a client queries for a simple hostname like `laptop`, Pocket Concierge will:

1. First try to resolve `laptop` exactly as specified
2. If not found, append the home domain and try `laptop.home`

### Host Mappings

Define custom hostname-to-IP mappings for your local network.

#### Host Entry Structure

```yaml
hosts:
  - hostname: "device-name"     # Required: hostname to resolve
    ipv4:                       # Optional: IPv4 addresses
      - "192.168.1.100"
      - "192.168.1.101"         # Multiple IPs for load balancing
    ipv6:                       # Optional: IPv6 addresses
      - "fe80::1234:5678:90ab:cdef"
      - "2001:db8::1"
```

#### Hostname Rules

- **Simple hostnames**: Names without dots get the home domain appended
  - `desktop` becomes `desktop.home`
- **Qualified hostnames**: Names with dots are used as-is
  - `nas.local` remains `nas.local`
- **Case sensitivity**: Hostnames are case-insensitive
- **Special characters**: Only alphanumeric characters, hyphens, and dots allowed

#### IP Address Rules

- **IPv4**: Standard dotted decimal notation (`192.168.1.100`)
- **IPv6**: Standard colon notation (`fe80::1234:5678:90ab:cdef`)
- **Multiple addresses**: DNS responses will include all configured addresses
- **Load balancing**: Clients will typically try addresses in order

#### Common Host Mapping Patterns

**Home devices:**

```yaml
hosts:
  - hostname: "laptop"
    ipv4: ["192.168.1.101"]
  
  - hostname: "phone"
    ipv4: ["192.168.1.102"]
  
  - hostname: "tablet"
    ipv4: ["192.168.1.103"]
```

**Servers with redundancy:**

```yaml
hosts:
  - hostname: "homeserver"
    ipv4:
      - "192.168.1.10"          # Primary interface
      - "192.168.2.10"          # Secondary interface
    ipv6:
      - "fd00::10"
```

**IoT devices:**

```yaml
hosts:
  - hostname: "thermostat"
    ipv4: ["192.168.1.75"]
  
  - hostname: "doorbell"
    ipv4: ["192.168.1.76"]
  
  - hostname: "security-camera-1"
    ipv4: ["192.168.1.80"]
  
  - hostname: "security-camera-2"
    ipv4: ["192.168.1.81"]
```

**Network infrastructure:**

```yaml
hosts:
  - hostname: "router"
    ipv4: ["192.168.1.1"]
  
  - hostname: "switch"
    ipv4: ["192.168.1.2"]
  
  - hostname: "ap1"
    ipv4: ["192.168.1.3"]
  
  - hostname: "ap2"
    ipv4: ["192.168.1.4"]
```

## Configuration Validation

Pocket Concierge validates the configuration at startup and reports errors for:

- **Invalid IP addresses**: Malformed IPv4 or IPv6 addresses
- **Invalid ports**: Port numbers outside valid range (1-65535)
- **Missing required fields**: Required parameters not specified
- **Protocol mismatches**: Invalid protocol/port combinations
- **DNS name conflicts**: Duplicate hostname entries

## Environment-Specific Configurations

### Development Environment

```yaml
server:
  port: 8053
  address: "127.0.0.1"

dns:
  cache_size: 1000
  ttl: 60                       # Short TTL for testing

upstream:
  - name: "Cloudflare"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

log_level: "debug"              # Verbose logging
home_dns_domain: "test"

hosts:
  - hostname: "dev-server"
    ipv4: ["127.0.0.1"]
```

### Production Home Network

```yaml
server:
  port: 8053
  address: "0.0.0.0"           # Listen on all interfaces

dns:
  cache_size: 50000            # Large cache
  ttl: 300

upstream:
  # Primary encrypted DNS
  - name: "Cloudflare DoH"
    address: "cloudflare-dns.com"
    protocol: "https"
    port: 443
    path: "/dns-query"
    verify: true
  
  # Fallback
  - name: "Cloudflare UDP"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

log_level: "info"
home_dns_domain: "home"

hosts:
  # Include all home devices
```

### High-Security Environment

```yaml
server:
  port: 8053
  address: "127.0.0.1"         # Local only

dns:
  enable_recursion: false      # Authoritative only
  cache_size: 10000
  ttl: 300

upstream:
  # Only encrypted DNS
  - name: "Quad9 DoT"
    address: "dns.quad9.net"
    protocol: "tls"
    port: 853
    verify: true

log_level: "warn"              # Minimal logging
home_dns_domain: "secure"
```

## Best Practices

### Security

1. **Use encrypted DNS**: Prefer DoH or DoT upstream servers
2. **Verify certificates**: Always set `verify: true` for encrypted protocols
3. **Limit binding**: Use `127.0.0.1` unless you need network-wide access
4. **Log monitoring**: Use appropriate log levels for your security requirements

### Performance

1. **Cache sizing**: Set cache size based on your network size
   - Small network (< 10 devices): 1,000 entries
   - Medium network (10-50 devices): 10,000 entries
   - Large network (> 50 devices): 50,000+ entries

2. **Upstream order**: List fastest/most reliable servers first
3. **Multiple protocols**: Include fallback servers with different protocols

### Reliability

1. **Multiple upstreams**: Configure at least 2-3 upstream servers
2. **Protocol diversity**: Mix encrypted and traditional DNS for fallback
3. **Regular testing**: Use the built-in test tools to verify configuration

### Maintenance

1. **Regular updates**: Keep upstream server configurations current
2. **Host management**: Use consistent naming conventions for hosts
3. **Documentation**: Comment complex configurations for future reference
