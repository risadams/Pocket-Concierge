# Deployment Guide

This guide covers various deployment scenarios for Pocket Concierge DNS server, from simple home setups to more advanced configurations.

## Quick Start Deployment

### Prerequisites

- Go 1.24 or later
- Network access to configure DNS settings
- Administrative privileges (for port 53 or system service installation)

### Basic Installation

1. **Download or build Pocket Concierge:**

   ```bash
   # Option 1: Build from source
   git clone https://github.com/risadams/Pocket-Concierge.git
   cd Pocket-Concierge
   make build
   
   # Option 2: Download pre-built binary (when available)
   # wget https://github.com/risadams/Pocket-Concierge/releases/latest/download/pocketconcierge
   ```

2. **Create configuration:**

   ```bash
   cp configs/example.yaml config.yaml
   # Edit config.yaml for your network
   ```

3. **Test the configuration:**

   ```bash
   ./build/pocketconcierge
   # Test with: go run test-dns.go desktop.home 127.0.0.1:8053
   ```

## Deployment Scenarios

### Scenario 1: Personal Computer DNS Override

Run Pocket Concierge on your personal computer and configure only that computer to use it.

**Configuration:**

```yaml
server:
  port: 8053
  address: "127.0.0.1"  # Only accept local connections

dns:
  ttl: 300
  enable_recursion: true
  cache_size: 5000

upstream:
  - name: "Cloudflare"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

home_dns_domain: "home"

hosts:
  - hostname: "laptop"
    ipv4: ["127.0.0.1"]  # Point to localhost
```

**Client Configuration:**

- **Windows**: Control Panel → Network → Change adapter settings → Properties → IPv4 → Use DNS: `127.0.0.1`
- **macOS**: System Preferences → Network → Advanced → DNS → Add `127.0.0.1`
- **Linux**: Edit `/etc/resolv.conf` or use NetworkManager

### Scenario 2: Home Router Integration

Install Pocket Concierge on your router or a dedicated device that serves your entire network.

**Configuration:**

```yaml
server:
  port: 53           # Standard DNS port
  address: "0.0.0.0" # Accept connections from entire network

dns:
  ttl: 300
  enable_recursion: true
  cache_size: 20000  # Larger cache for multiple devices

upstream:
  # Use encrypted DNS for privacy
  - name: "Cloudflare DoH"
    address: "cloudflare-dns.com"
    protocol: "https"
    port: 443
    path: "/dns-query"
    verify: true
  
  # Fallback to traditional DNS
  - name: "Cloudflare UDP"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

home_dns_domain: "home"

hosts:
  - hostname: "router"
    ipv4: ["192.168.1.1"]
  
  - hostname: "nas"
    ipv4: ["192.168.1.10"]
  
  - hostname: "printer"
    ipv4: ["192.168.1.53"]
```

**Router Configuration:**

- Set DHCP to provide Pocket Concierge server IP as primary DNS
- Ensure firewall allows DNS traffic on port 53
- Consider running as system service for reliability

### Scenario 3: Raspberry Pi Dedicated DNS Server

Deploy on a Raspberry Pi as a dedicated DNS appliance for your home network.

**Hardware Requirements:**

- Raspberry Pi 3B+ or newer
- MicroSD card (16GB minimum)
- Stable power supply
- Network connection (Ethernet recommended)

**Software Setup:**

```bash
# Install Go on Raspberry Pi
sudo apt update
sudo apt install golang-go

# Build Pocket Concierge
git clone https://github.com/risadams/Pocket-Concierge.git
cd Pocket-Concierge
go build -o pocketconcierge ./cmd/pocketconcierge

# Create system user
sudo useradd -r -s /bin/false pocketconcierge

# Install binary
sudo cp pocketconcierge /usr/local/bin/
sudo chmod +x /usr/local/bin/pocketconcierge

# Create configuration directory
sudo mkdir -p /etc/pocketconcierge
sudo cp config.yaml /etc/pocketconcierge/
```

**Configuration for Pi:**

```yaml
server:
  port: 53
  address: "0.0.0.0"

dns:
  ttl: 300
  enable_recursion: true
  cache_size: 50000  # Pi has plenty of RAM

upstream:
  # Multiple encrypted upstreams for reliability
  - name: "Cloudflare DoT"
    address: "1dot1dot1dot1.cloudflare-dns.com"
    protocol: "tls"
    port: 853
    verify: true
  
  - name: "Quad9 DoT"
    address: "dns.quad9.net"
    protocol: "tls"
    port: 853
    verify: true
  
  - name: "Cloudflare Fallback"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53

log_level: "info"
home_dns_domain: "home"

# Add all your home devices here
hosts:
  - hostname: "pi"
    ipv4: ["192.168.1.100"]
  # ... more hosts
```

### Scenario 4: Docker Container Deployment

Run Pocket Concierge in a Docker container for easy deployment and management.

**Dockerfile:**

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o pocketconcierge ./cmd/pocketconcierge

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/pocketconcierge .
COPY --from=builder /app/configs/example.yaml ./config.yaml

EXPOSE 53/udp 53/tcp 8053/udp 8053/tcp

CMD ["./pocketconcierge", "config.yaml"]
```

**Build and run:**

```bash
# Build container
docker build -t pocketconcierge .

# Run container
docker run -d \
  --name pocketconcierge \
  -p 8053:8053/udp \
  -p 8053:8053/tcp \
  -v /path/to/your/config.yaml:/root/config.yaml:ro \
  pocketconcierge

# For standard DNS port (requires privileged)
docker run -d \
  --name pocketconcierge \
  --privileged \
  -p 53:53/udp \
  -p 53:53/tcp \
  -v /path/to/your/config.yaml:/root/config.yaml:ro \
  pocketconcierge
```

**Docker Compose:**

```yaml
version: '3.8'

services:
  pocketconcierge:
    build: .
    container_name: pocketconcierge
    restart: unless-stopped
    ports:
      - "8053:8053/udp"
      - "8053:8053/tcp"
    volumes:
      - ./config.yaml:/root/config.yaml:ro
    networks:
      - dns-network

networks:
  dns-network:
    driver: bridge
```

## System Service Installation

### Linux (systemd)

Create a systemd service for automatic startup and management.

**Service file (`/etc/systemd/system/pocketconcierge.service`):**

```ini
[Unit]
Description=Pocket Concierge DNS Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=pocketconcierge
Group=pocketconcierge
WorkingDirectory=/etc/pocketconcierge
ExecStart=/usr/local/bin/pocketconcierge /etc/pocketconcierge/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/pocketconcierge
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

**Installation commands:**

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable pocketconcierge

# Start service
sudo systemctl start pocketconcierge

# Check status
sudo systemctl status pocketconcierge

# View logs
sudo journalctl -u pocketconcierge -f
```

### Windows Service

Use a tool like NSSM (Non-Sucking Service Manager) to create a Windows service.

**Installation:**

```powershell
# Download NSSM from https://nssm.cc/

# Install service
nssm install PocketConcierge "C:\path\to\pocketconcierge.exe"
nssm set PocketConcierge AppParameters "C:\path\to\config.yaml"
nssm set PocketConcierge DisplayName "Pocket Concierge DNS Server"
nssm set PocketConcierge Description "Home network DNS server"
nssm set PocketConcierge Start SERVICE_AUTO_START

# Start service
nssm start PocketConcierge
```

### macOS (launchd)

Create a launchd plist for macOS service management.

**Service file (`/Library/LaunchDaemons/com.pocketconcierge.dns.plist`):**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.pocketconcierge.dns</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/pocketconcierge</string>
        <string>/usr/local/etc/pocketconcierge/config.yaml</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>UserName</key>
    <string>_pocketconcierge</string>
    
    <key>StandardOutPath</key>
    <string>/var/log/pocketconcierge.log</string>
    
    <key>StandardErrorPath</key>
    <string>/var/log/pocketconcierge.error.log</string>
</dict>
</plist>
```

**Installation:**

```bash
# Load service
sudo launchctl load /Library/LaunchDaemons/com.pocketconcierge.dns.plist

# Start service
sudo launchctl start com.pocketconcierge.dns

# Check status
sudo launchctl list | grep pocketconcierge
```

## Network Configuration

### DHCP Configuration

Configure your DHCP server to automatically provide clients with Pocket Concierge as their DNS server.

**Router/DHCP Server Settings:**

- Primary DNS: `<pocket-concierge-ip>`
- Secondary DNS: `1.1.1.1` (fallback)

**Common Router Interfaces:**

**DD-WRT/OpenWrt:**

```
# In DHCP settings
Static DNS 1: 192.168.1.100  # Pocket Concierge IP
Static DNS 2: 1.1.1.1        # Fallback
```

**pfSense:**

```
Services → DHCP Server → DNS Servers
192.168.1.100  # Pocket Concierge
1.1.1.1       # Fallback
```

**Unifi:**

```
Networks → LAN → DHCP → DNS Server
192.168.1.100  # Pocket Concierge
1.1.1.1       # Fallback
```

### Firewall Configuration

Ensure appropriate firewall rules allow DNS traffic.

**Linux (iptables):**

```bash
# Allow DNS queries
sudo iptables -A INPUT -p udp --dport 53 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 53 -j ACCEPT

# For non-standard port
sudo iptables -A INPUT -p udp --dport 8053 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8053 -j ACCEPT

# Save rules
sudo iptables-save > /etc/iptables/rules.v4
```

**Windows Firewall:**

```powershell
# Allow DNS through Windows Firewall
New-NetFirewallRule -DisplayName "Pocket Concierge DNS" -Direction Inbound -Protocol UDP -LocalPort 53 -Action Allow
New-NetFirewallRule -DisplayName "Pocket Concierge DNS TCP" -Direction Inbound -Protocol TCP -LocalPort 53 -Action Allow
```

## Monitoring and Maintenance

### Log Management

**Configure log rotation:**

```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/pocketconcierge << EOF
/var/log/pocketconcierge.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload pocketconcierge
    endscript
}
EOF
```

### Health Monitoring

**Basic health check script:**

```bash
#!/bin/bash
# health-check.sh

DNS_SERVER="127.0.0.1:8053"
TEST_HOSTNAME="google.com"

if dig @${DNS_SERVER} ${TEST_HOSTNAME} +short > /dev/null; then
    echo "✅ Pocket Concierge is healthy"
    exit 0
else
    echo "❌ Pocket Concierge health check failed"
    exit 1
fi
```

**Automated monitoring with cron:**

```bash
# Add to crontab (crontab -e)
*/5 * * * * /path/to/health-check.sh || systemctl restart pocketconcierge
```

### Performance Monitoring

**Monitor using built-in tools:**

```bash
# Benchmark performance
go run cmd/benchmark/main.go

# Load testing
go run cmd/loadtest/main.go
```

**System resource monitoring:**

```bash
# Check memory usage
ps aux | grep pocketconcierge

# Check network connections
netstat -tulpn | grep :53

# Monitor DNS queries (if logging enabled)
tail -f /var/log/pocketconcierge.log | grep "query"
```

## Troubleshooting

### Common Issues

**1. Permission denied on port 53:**

```bash
# Solution: Run with sudo or use capability
sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/pocketconcierge
```

**2. Configuration not loading:**

```bash
# Check file permissions
ls -la config.yaml

# Validate YAML syntax
go run -c 'import yaml; yaml.safe_load(open("config.yaml"))'
```

**3. Upstream DNS failures:**

```bash
# Test upstream connectivity
dig @1.1.1.1 google.com

# Check firewall/network connectivity
telnet 1.1.1.1 53
```

**4. High memory usage:**

```bash
# Reduce cache size in config
dns:
  cache_size: 1000  # Reduce from default
```

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
log_level: "debug"
```

### Network Testing

**Test resolution from different clients:**

```bash
# From another machine on the network
dig @<pocket-concierge-ip> desktop.home

# Test IPv6
dig @<pocket-concierge-ip> desktop.home AAAA

# Test recursion
dig @<pocket-concierge-ip> google.com
```

## Security Considerations

### Network Security

- **Bind only to necessary interfaces**: Use `127.0.0.1` for local-only access
- **Firewall rules**: Only allow DNS traffic from trusted networks
- **Regular updates**: Keep Pocket Concierge updated with latest security fixes

### Configuration Security

- **File permissions**: Restrict config file access (`chmod 600 config.yaml`)
- **User privileges**: Run as dedicated non-privileged user
- **Upstream encryption**: Prefer DoT/DoH upstream servers

### Monitoring Security

- **Log monitoring**: Monitor for unusual query patterns
- **Rate limiting**: Consider implementing rate limiting for production use
- **Access logs**: Enable logging to track DNS usage patterns

## Backup and Recovery

### Configuration Backup

```bash
# Backup configuration
cp /etc/pocketconcierge/config.yaml /backup/pocketconcierge-config-$(date +%Y%m%d).yaml

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/pocketconcierge"
DATE=$(date +%Y%m%d)
mkdir -p $BACKUP_DIR
cp /etc/pocketconcierge/config.yaml $BACKUP_DIR/config-$DATE.yaml
```

### Disaster Recovery

```bash
# Quick recovery procedure
1. Install Pocket Concierge binary
2. Restore configuration file
3. Start service
4. Update DHCP/network settings
5. Test resolution
```

### High Availability Setup

For critical deployments, consider:

- **Multiple instances**: Run on different hardware
- **Load balancer**: Use DNS load balancing
- **Monitoring**: Automated failover monitoring
- **Backup DNS**: Always configure fallback DNS servers
