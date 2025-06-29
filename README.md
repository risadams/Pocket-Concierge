# 🏨 PocketConcierge

A tiny DNS server for home networks - lightweight, cross-platform, and easy to configure.

PocketConcierge acts as your home network's personal concierge, helping devices find each other by name instead of remembering IP addresses. Perfect for home labs, small offices, or anyone who wants friendly hostnames for their local devices.

## ✨ Features

- **🏠 Local Host Resolution** - Map custom hostnames to IP addresses
- **🔄 Upstream Forwarding** - Forward unknown queries to your favorite DNS providers
- **⚡ Fast & Lightweight** - Written in Go with minimal resource usage
- **🔧 Easy Configuration** - Simple YAML configuration file
- **🌐 IPv4 & IPv6 Support** - Handle both address types seamlessly
- **🎯 Cross-Platform** - Runs on Windows, macOS, and Linux
- **📊 Built-in Logging** - Monitor queries and responses
- **🛡️ Graceful Shutdown** - Clean exit with proper signal handling

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or later
- Admin/root privileges (for binding to port 53)

### Run with default settings

```bash
go run cmd/pocketconcierge/main.go
```

### Run with custom config

```bash
go run cmd/pocketconcierge/main.go /path/to/config.yaml
```

### Build and install

```bash
# Build for your platform
go build -o pocketconcierge cmd/pocketconcierge/main.go

# Run the binary
./pocketconcierge
```

## ⚙️ Configuration

PocketConcierge uses a YAML configuration file. If no config file is specified, it looks for `config.yaml` in the current directory.

### Example Configuration

```yaml
# Server settings
server:
  port: 53              # DNS port (use 5353 for testing without admin rights)
  address: "0.0.0.0"    # Bind address (0.0.0.0 = all interfaces)

# DNS behavior
dns:
  ttl: 300              # Default TTL for responses (seconds)
  enable_recursion: true # Forward unknown queries upstream
  cache_size: 1000      # Response cache size

# Upstream DNS servers
upstream:
  - "8.8.8.8:53"        # Google DNS
  - "1.1.1.1:53"        # Cloudflare DNS

# Logging level: debug, info, warn, error
log_level: "info"

# Local hostname mappings
hosts:
  - hostname: "homeserver.home"
    ipv4:
      - "192.168.1.50"
  
  - hostname: "laptop.home"
    ipv4:
      - "192.168.1.101"
    ipv6:
      - "fe80::1234:5678:90ab:cdef"
  
  - hostname: "router.home"
    ipv4:
      - "192.168.1.1"
```

### Configuration Options

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| `server.port` | Port number | DNS server port | `53` |
| `server.address` | Bind address | Interface to bind to | `0.0.0.0` |
| `dns.ttl` | TTL seconds | Time-to-live for responses | `300` |
| `dns.enable_recursion` | Boolean | Forward unknown queries | `true` |
| `dns.cache_size` | Number | Response cache size | `1000` |
| `upstream` | Array | Upstream DNS servers | Google & Cloudflare |
| `log_level` | String | Logging verbosity | `info` |
| `hosts` | Array | Local hostname mappings | `[]` |

## 🏠 Setting Up Local Hostnames

Add entries to the `hosts` section to create custom local DNS mappings:

```yaml
hosts:
  # Simple mapping
  - hostname: "myserver.home"
    ipv4:
      - "192.168.1.100"
  
  # Multiple IPs (load balancing/failover)
  - hostname: "cluster.home"
    ipv4:
      - "192.168.1.10"
      - "192.168.1.11"
      - "192.168.1.12"
  
  # IPv6 support
  - hostname: "modern.home"
    ipv4:
      - "192.168.1.200"
    ipv6:
      - "2001:db8::1"
```

## 🔧 Common Use Cases

### Home Lab Setup

Perfect for naming your home lab servers:

```yaml
hosts:
  - hostname: "proxmox.lab"
    ipv4: ["192.168.1.10"]
  - hostname: "truenas.lab"
    ipv4: ["192.168.1.20"]
  - hostname: "docker.lab"
    ipv4: ["192.168.1.30"]
```

### Development Environment

Great for local development:

```yaml
hosts:
  - hostname: "api.dev"
    ipv4: ["127.0.0.1"]
  - hostname: "frontend.dev"
    ipv4: ["127.0.0.1"]
  - hostname: "db.dev"
    ipv4: ["192.168.1.100"]
```

### IoT Device Naming

Name your smart home devices:

```yaml
hosts:
  - hostname: "camera-garage.iot"
    ipv4: ["192.168.1.150"]
  - hostname: "sensor-living.iot"
    ipv4: ["192.168.1.151"]
  - hostname: "hub-main.iot"
    ipv4: ["192.168.1.152"]
```

## 🛠️ Installation & Deployment

### Testing Setup (Non-privileged)

For testing without admin rights, use a non-standard port:

```yaml
server:
  port: 5353  # Non-privileged port
  address: "127.0.0.1"
```

Test with:

```bash
# Query the local server
dig @127.0.0.1 -p 5353 homeserver.home
```

### Production Setup

#### Linux (systemd)

Step-by-step installation:

1. Build the binary:

```bash
go build -o /usr/local/bin/pocketconcierge cmd/pocketconcierge/main.go
```

1. Create config directory:

```bash
sudo mkdir -p /etc/pocketconcierge
sudo cp configs/example.yaml /etc/pocketconcierge/config.yaml
```

1. Create systemd service:

```bash
sudo tee /etc/systemd/system/pocketconcierge.service << EOF
[Unit]
Description=PocketConcierge DNS Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/pocketconcierge /etc/pocketconcierge/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

1. Enable and start:

```bash
sudo systemctl enable pocketconcierge
sudo systemctl start pocketconcierge
```

#### Windows (Service)

Use a tool like NSSM (Non-Sucking Service Manager) to run as a Windows service.

#### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o pocketconcierge cmd/pocketconcierge/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pocketconcierge .
COPY configs/example.yaml config.yaml
EXPOSE 53/udp
CMD ["./pocketconcierge"]
```

## 🔍 Troubleshooting

### Permission Denied (Port 53)

DNS typically runs on port 53, which requires admin privileges:

**Linux/macOS:**

```bash
sudo go run cmd/pocketconcierge/main.go
```

**Windows:**
Run PowerShell as Administrator

**Alternative:** Use port 5353 for testing:

```yaml
server:
  port: 5353
```

### Testing DNS Resolution

```bash
# Test A record
dig @your-server-ip homeserver.home

# Test specific port
dig @127.0.0.1 -p 5353 homeserver.home

# Test with nslookup
nslookup homeserver.home your-server-ip
```

### Common Issues

| Problem | Solution |
|---------|----------|
| "Permission denied" on port 53 | Run with admin/root or use port 5353 |
| "No such host" responses | Check hostname spelling in config |
| Upstream queries failing | Verify internet connectivity and upstream DNS |
| Service won't start | Check config file syntax with `yaml` validator |

## 📊 Monitoring

PocketConcierge logs all DNS queries and responses. Log levels available:

- `debug` - Verbose logging including query details
- `info` - Standard operational logs (default)
- `warn` - Warning messages only
- `error` - Error messages only

Example log output:

```text
🏨 PocketConcierge DNS Server v0.1.0
✅ Loaded configuration from config.yaml
🚀 Starting DNS server on 0.0.0.0:53
🔍 Query: A homeserver.home.
✅ Local resolve: homeserver.home. -> 1 answers
🔍 Query: A google.com.
🔄 Upstream resolve: google.com. -> 1 answers
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

### Development Setup

```bash
git clone https://github.com/risadams/Pocket-Concierge.git
cd Pocket-Concierge
go mod tidy
go run cmd/pocketconcierge/main.go configs/example.yaml
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with the excellent [miekg/dns](https://github.com/miekg/dns) Go library
- Inspired by the need for simple home network DNS solutions
