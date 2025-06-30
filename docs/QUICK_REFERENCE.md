# Quick Reference

This is a quick reference card for common Pocket Concierge operations and configurations.

## Essential Commands

### Start/Stop Server

```bash
# Start server with default config
./pocketconcierge

# Start with custom config
./pocketconcierge /path/to/config.yaml

# Start as systemd service
sudo systemctl start pocketconcierge

# Stop service
sudo systemctl stop pocketconcierge
```

### Testing DNS Resolution

```bash
# Test local hostname
dig @127.0.0.1 -p 8053 desktop.home

# Test external domain
dig @127.0.0.1 -p 8053 google.com

# Test with built-in tool
go run test-dns.go desktop.home 127.0.0.1:8053

# Test IPv6
dig @127.0.0.1 -p 8053 server.home AAAA
```

### Performance Testing

```bash
# Run benchmark
go run cmd/benchmark/main.go

# Run load test
go run cmd/loadtest/main.go
```

## Configuration Quick Start

### Minimal Configuration

```yaml
server:
  port: 8053
  address: "127.0.0.1"

upstream:
  - name: "Cloudflare"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

home_dns_domain: "home"

hosts:
  - hostname: "laptop"
    ipv4: ["192.168.1.100"]
```

### Production Configuration

```yaml
server:
  port: 53
  address: "0.0.0.0"

dns:
  ttl: 300
  enable_recursion: true
  cache_size: 10000

upstream:
  # Primary encrypted
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
  - hostname: "router"
    ipv4: ["192.168.1.1"]
  - hostname: "server"
    ipv4: ["192.168.1.10"]
  - hostname: "nas"
    ipv4: ["192.168.1.20"]
```

## Common Host Configurations

### Single IP Address

```yaml
hosts:
  - hostname: "desktop"
    ipv4: ["192.168.1.100"]
```

### Multiple IP Addresses (Load Balancing)

```yaml
hosts:
  - hostname: "webserver"
    ipv4:
      - "192.168.1.10"
      - "192.168.1.11"
      - "192.168.1.12"
```

### IPv4 and IPv6

```yaml
hosts:
  - hostname: "server"
    ipv4: ["192.168.1.10"]
    ipv6: ["2001:db8::10"]
```

### Full Domain Names

```yaml
hosts:
  - hostname: "nas.local"  # Won't get .home suffix
    ipv4: ["192.168.1.20"]
```

## Popular Upstream Servers

### Cloudflare

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

### Google Public DNS

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

### Quad9

```yaml
# DNS-over-TLS
- name: "Quad9 DoT"
  address: "dns.quad9.net"
  protocol: "tls"
  port: 853
  verify: true

# Traditional DNS
- name: "Quad9 UDP"
  address: "9.9.9.9"
  protocol: "udp"
  port: 53
```

## Troubleshooting Quick Fixes

### Server Won't Start

```bash
# Check if port is in use
sudo netstat -tulpn | grep :53

# Try different port
# Edit config.yaml: server.port: 8053

# Check permissions for port 53
sudo setcap CAP_NET_BIND_SERVICE=+eip ./pocketconcierge
```

### DNS Not Resolving

```bash
# Check server is running
ps aux | grep pocketconcierge

# Test direct connection
dig @127.0.0.1 -p 8053 google.com

# Check client DNS settings
cat /etc/resolv.conf  # Linux
nslookup  # Windows
```

### Performance Issues

```yaml
# Reduce cache size
dns:
  cache_size: 1000

# Use faster upstream
upstream:
  - name: "Local Router"
    address: "192.168.1.1"
    protocol: "udp"
    port: 53
```

## Log Level Reference

| Level | Purpose | Use When |
|-------|---------|----------|
| `debug` | Detailed debugging | Troubleshooting issues |
| `info` | General information | Normal operation |
| `warn` | Warning messages | Production monitoring |
| `error` | Error messages only | Minimal logging |

## Port Reference

| Port | Protocol | Purpose | Privileges |
|------|----------|---------|------------|
| 53 | DNS | Standard DNS port | Requires root/admin |
| 8053 | DNS | Default Pocket Concierge | No special privileges |
| 443 | HTTPS | DNS-over-HTTPS | N/A (client) |
| 853 | TLS | DNS-over-TLS | N/A (client) |

## File Locations

### Default Locations

| File | Default Location | Purpose |
|------|------------------|---------|
| Configuration | `./config.yaml` | Server configuration |
| Binary | `./pocketconcierge` | Executable |
| Logs | stdout/stderr | Application logs |

### System Service Locations

| File | Location | Purpose |
|------|----------|---------|
| Binary | `/usr/local/bin/pocketconcierge` | System executable |
| Configuration | `/etc/pocketconcierge/config.yaml` | System configuration |
| Service | `/etc/systemd/system/pocketconcierge.service` | Systemd service |
| Logs | `journalctl -u pocketconcierge` | System logs |

## Environment Variables

Pocket Concierge doesn't use environment variables by default, but you can set them in systemd service files:

```ini
[Service]
Environment="LOG_LEVEL=debug"
Environment="CONFIG_PATH=/etc/pocketconcierge/config.yaml"
```

## System Service Commands

### Systemd (Linux)

```bash
# Install service
sudo systemctl enable pocketconcierge

# Start service
sudo systemctl start pocketconcierge

# Stop service
sudo systemctl stop pocketconcierge

# Restart service
sudo systemctl restart pocketconcierge

# Check status
sudo systemctl status pocketconcierge

# View logs
sudo journalctl -u pocketconcierge -f
```

### Docker

```bash
# Run container
docker run -d --name pocketconcierge \
  -p 8053:8053/udp \
  -v ./config.yaml:/app/config.yaml \
  pocketconcierge

# Check status
docker ps | grep pocketconcierge

# View logs
docker logs pocketconcierge

# Stop container
docker stop pocketconcierge
```

## Client Configuration

### Set DNS Server

**Windows:**

```cmd
netsh interface ip set dns "Local Area Connection" static 192.168.1.100
```

**macOS:**

```bash
sudo networksetup -setdnsservers "Ethernet" 192.168.1.100
```

**Linux:**

```bash
# Edit /etc/resolv.conf
echo "nameserver 192.168.1.100" | sudo tee /etc/resolv.conf
```

### Test DNS Settings

**Windows:**

```cmd
ipconfig /all | findstr "DNS Servers"
nslookup desktop.home
```

**macOS/Linux:**

```bash
cat /etc/resolv.conf
dig desktop.home
```

## Performance Tuning

### For Home Networks (< 10 devices)

```yaml
dns:
  cache_size: 1000
  ttl: 300
```

### For Larger Networks (10-50 devices)

```yaml
dns:
  cache_size: 10000
  ttl: 300
```

### For High-Performance Networks

```yaml
dns:
  cache_size: 50000
  ttl: 600
```

## Security Hardening

### Minimal Permissions

```yaml
server:
  address: "127.0.0.1"  # Local only
```

### Encrypted Upstream Only

```yaml
upstream:
  - name: "Cloudflare DoT"
    address: "1dot1dot1dot1.cloudflare-dns.com"
    protocol: "tls"
    port: 853
    verify: true
```

### File Permissions

```bash
chmod 600 config.yaml
chown pocketconcierge:pocketconcierge config.yaml
```

---

**For more detailed information, see the full documentation:**

- [Configuration Guide](CONFIGURATION.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
