# Frequently Asked Questions (FAQ)

## General Questions

### What is Pocket Concierge?

Pocket Concierge is a lightweight DNS server designed specifically for home networks. It provides local hostname resolution (like `laptop.home`), DNS caching for improved performance, and secure upstream DNS forwarding with support for DNS-over-HTTPS (DoH) and DNS-over-TLS (DoT).

### Why would I need a local DNS server?

A local DNS server provides several benefits:

- **Custom hostnames**: Access your devices with memorable names like `server.home` instead of IP addresses
- **Faster DNS resolution**: Local caching reduces response times for frequently accessed domains
- **Enhanced privacy**: Use encrypted DNS (DoH/DoT) to prevent ISP DNS monitoring
- **Network control**: Block unwanted domains or redirect traffic as needed
- **Offline resolution**: Local hostnames work even without internet connectivity

### How is Pocket Concierge different from other DNS servers?

Pocket Concierge is specifically designed for home networks with features like:

- **Simple configuration**: Easy YAML configuration file
- **Home domain support**: Automatic `.home` suffix for simple hostnames
- **Multiple protocols**: Support for UDP, TCP, DoH, and DoT upstream servers
- **Lightweight**: Minimal resource usage perfect for Raspberry Pi or router deployment
- **Built-in tools**: Benchmarking and load testing utilities included

### Is Pocket Concierge secure?

Yes, Pocket Concierge includes several security features:

- **Encrypted DNS**: Support for DNS-over-HTTPS and DNS-over-TLS
- **Certificate verification**: Validates upstream server certificates
- **Query validation**: Sanitizes and validates incoming DNS queries
- **Local network focus**: Designed for trusted home network environments

## Installation and Setup

### What are the system requirements?

- **Operating System**: Windows, macOS, Linux, or any Go-supported platform
- **Go Version**: Go 1.24 or later (for building from source)
- **Memory**: Minimum 16MB RAM, recommended 64MB+
- **Storage**: 10MB for binary and configuration
- **Network**: Access to upstream DNS servers

### Can I run Pocket Concierge on a Raspberry Pi?

Yes! Pocket Concierge works great on Raspberry Pi. It's lightweight enough to run on:

- Raspberry Pi Zero W (512MB RAM)
- Raspberry Pi 3B+ or newer (recommended)
- Other ARM-based single-board computers

See the [Deployment Guide](DEPLOYMENT.md) for detailed Raspberry Pi setup instructions.

### Do I need root/administrator privileges?

It depends on your configuration:

- **Port 8053 (default)**: No special privileges needed
- **Port 53 (standard DNS)**: Requires root/administrator or special capabilities
- **System service**: Usually requires administrative setup

You can use `setcap` on Linux to grant port binding privileges without running as root.

### Can I run multiple instances?

Yes, but each instance needs:

- **Different port numbers**: Each instance must bind to a unique port
- **Separate configuration files**: Different config files for each instance
- **Distinct purposes**: Consider why you need multiple instances

Common scenarios for multiple instances:

- Testing configuration changes
- Separate internal/external DNS handling
- Development vs. production environments

## Configuration

### How do I configure local hostnames?

Add hostnames to the `hosts` section of your configuration:

```yaml
hosts:
  - hostname: "laptop"
    ipv4:
      - "192.168.1.100"
  
  - hostname: "server"
    ipv4:
      - "192.168.1.10"
    ipv6:
      - "fe80::1234:5678:90ab:cdef"
```

Simple hostnames automatically get the home domain suffix (default `.home`).

### What upstream DNS servers should I use?

Popular options include:

**For privacy (encrypted):**

- Cloudflare DoT: `1dot1dot1dot1.cloudflare-dns.com:853`
- Quad9 DoT: `dns.quad9.net:853`
- Cloudflare DoH: `cloudflare-dns.com/dns-query`

**For performance (unencrypted):**

- Cloudflare: `1.1.1.1`
- Google: `8.8.8.8`
- Your ISP's DNS servers

**For filtering:**

- OpenDNS: `208.67.222.222` (blocks malware)
- Quad9: `9.9.9.9` (blocks malicious domains)

### How do I change the default port?

Edit the `server` section in your configuration:

```yaml
server:
  port: 53        # Use standard DNS port
  address: "0.0.0.0"  # Listen on all interfaces
```

Remember that port 53 requires elevated privileges.

### Can I use wildcard hostnames?

Currently, Pocket Concierge doesn't support wildcard hostnames. Each hostname must be explicitly configured. This is a planned feature for future releases.

### How do I configure IPv6?

Add IPv6 addresses to your host configurations:

```yaml
hosts:
  - hostname: "server"
    ipv4:
      - "192.168.1.10"
    ipv6:
      - "2001:db8::10"
      - "fe80::1234:5678:90ab:cdef"
```

Clients can query for AAAA records to get IPv6 addresses.

## Network Integration

### How do I configure my router to use Pocket Concierge?

The process varies by router, but generally:

1. Access your router's admin interface
2. Find DHCP settings
3. Set DNS server to your Pocket Concierge IP address
4. Set a secondary DNS (like `1.1.1.1`) as fallback
5. Save and restart DHCP service

### Can I use Pocket Concierge with existing DNS servers?

Yes! You can configure Pocket Concierge as:

- **Primary DNS**: Handle local queries, forward external queries
- **Secondary DNS**: Backup for your primary DNS server
- **Selective DNS**: Only for specific devices or domains

### Will this work with my mesh network?

Yes, Pocket Concierge works with mesh networks. Deploy it on:

- **Main router**: Central DNS for entire mesh
- **Individual nodes**: Local DNS per mesh node
- **Dedicated device**: Separate device connected to mesh

### How do I test if it's working?

Use these commands to test:

```bash
# Test local hostname resolution
dig @<pocket-concierge-ip> desktop.home

# Test external domain resolution
dig @<pocket-concierge-ip> google.com

# Test from the built-in tool
go run test-dns.go desktop.home <pocket-concierge-ip>:8053
```

## Performance and Troubleshooting

### How much memory does Pocket Concierge use?

Typical memory usage:

- **Base usage**: ~10-20MB
- **With 10,000 cache entries**: ~30-40MB
- **Under heavy load**: ~50-100MB

Memory usage scales with cache size and query volume.

### How fast is DNS resolution?

Response times depend on the query type:

- **Cached local hostnames**: < 1ms
- **Cached external domains**: < 1ms
- **Uncached external domains**: 10-200ms (depends on upstream)

### Why are my DNS queries slow?

Common causes and solutions:

**Slow upstream servers:**

- Try different upstream DNS servers
- Use geographically closer servers
- Check network connectivity to upstream servers

**Large cache:**

- Reduce cache size if memory is limited
- Monitor cache hit rates

**Network issues:**

- Check network connectivity
- Verify firewall rules
- Test with simpler configuration

### How do I monitor performance?

Built-in tools:

```bash
# Benchmark DNS performance
go run cmd/benchmark/main.go

# Load testing
go run cmd/loadtest/main.go
```

System monitoring:

```bash
# Check process stats
ps aux | grep pocketconcierge

# Monitor network connections
netstat -tulpn | grep :53
```

### What if Pocket Concierge stops working?

Basic troubleshooting steps:

1. **Check if process is running**: `ps aux | grep pocketconcierge`
2. **Check logs**: `journalctl -u pocketconcierge` (systemd)
3. **Test configuration**: Reload with debug logging enabled
4. **Verify network**: Test upstream DNS connectivity
5. **Restart service**: `systemctl restart pocketconcierge`

See the [Troubleshooting Guide](TROUBLESHOOTING.md) for detailed help.

## Security and Privacy

### Is my DNS traffic encrypted?

It depends on your upstream configuration:

- **DoH/DoT upstream**: Queries to upstream servers are encrypted
- **Local queries**: Traffic between clients and Pocket Concierge is unencrypted
- **Traditional upstream**: Queries to upstream servers are unencrypted

For maximum privacy, use DoH or DoT upstream servers.

### Can I block unwanted domains?

Currently, Pocket Concierge doesn't include domain blocking features. This functionality could be added through:

- **Upstream DNS filtering**: Use filtering DNS services like OpenDNS
- **Custom logic**: Modify the source code to add blocking
- **External tools**: Use alongside tools like Pi-hole

### Should I be concerned about DNS logs?

Pocket Concierge logs are stored locally and not shared. Log content depends on your log level:

- **Error**: Only errors and critical issues
- **Info**: Basic operational information
- **Debug**: Detailed query and response information

Adjust the log level based on your privacy preferences.

### How do I secure the configuration file?

Protect your configuration file:

```bash
# Set restrictive permissions
chmod 600 config.yaml
chown pocketconcierge:pocketconcierge config.yaml

# For systemd service
sudo mkdir -p /etc/pocketconcierge
sudo cp config.yaml /etc/pocketconcierge/
sudo chmod 600 /etc/pocketconcierge/config.yaml
```

## Advanced Usage

### Can I use Pocket Concierge in Docker?

Yes! See the [Deployment Guide](DEPLOYMENT.md) for Docker configuration examples, including:

- Dockerfile for building containers
- Docker Compose configurations
- Volume mounting for configuration files

### How do I set up high availability?

For critical deployments:

1. **Multiple instances**: Run on different hardware
2. **Load balancing**: Use DNS round-robin or load balancer
3. **Monitoring**: Automated health checks and failover
4. **Backup DNS**: Always configure fallback DNS servers

### Can I integrate with monitoring systems?

Currently, Pocket Concierge provides:

- **Structured logging**: JSON or text logs
- **Built-in benchmarks**: Performance testing tools
- **System service integration**: Works with systemd, Docker, etc.

Future versions may include:

- Prometheus metrics endpoint
- Health check HTTP endpoint
- SNMP support

### How do I contribute to development?

Ways to contribute:

1. **Report bugs**: Use GitHub issues
2. **Request features**: Suggest improvements
3. **Submit code**: Pull requests welcome
4. **Documentation**: Help improve docs
5. **Testing**: Test on different platforms

See the [Contributing Guide](../CONTRIBUTING.md) for details.

## Migration and Compatibility

### Can I migrate from Pi-hole?

Yes, but they serve different purposes:

- **Pi-hole**: Ad blocking + DNS resolution
- **Pocket Concierge**: Local hostname resolution + DNS forwarding

You can:

- **Run both**: Pi-hole for filtering, Pocket Concierge for local hosts
- **Replace**: Use filtering upstream DNS instead of Pi-hole
- **Migrate configs**: Convert Pi-hole local DNS entries to Pocket Concierge hosts

### Is Pocket Concierge compatible with existing DNS setups?

Yes, Pocket Concierge integrates well with:

- **Existing routers**: Use as primary or secondary DNS
- **DHCP servers**: Configure to provide Pocket Concierge as DNS server
- **Network equipment**: Works with managed switches, firewalls, etc.
- **Operating systems**: Compatible with all major OS DNS client implementations

### How do I backup my configuration?

Simple backup strategy:

```bash
# Create backup
cp config.yaml config-backup-$(date +%Y%m%d).yaml

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/pocketconcierge"
mkdir -p $BACKUP_DIR
cp /etc/pocketconcierge/config.yaml $BACKUP_DIR/config-$(date +%Y%m%d).yaml
```

For complete backup, also save:

- Configuration file
- Any custom scripts
- Log files (if needed)
- Service configuration (systemd, etc.)

### Can I import DNS records from other sources?

Currently, DNS records must be manually configured in the YAML file. Potential import sources for future versions:

- **CSV files**: Bulk import from spreadsheets
- **DNS zone files**: Standard DNS zone format
- **Network discovery**: Automatic discovery of network devices
- **Router exports**: Import from router configuration

## Future Development

### What features are planned?

Potential future features include:

- **Web interface**: GUI for configuration and monitoring
- **API endpoints**: RESTful API for dynamic configuration
- **Domain blocking**: Built-in ad/malware blocking
- **Wildcard support**: Pattern-based hostname matching
- **Metrics export**: Prometheus/monitoring integration
- **Clustering**: Multi-instance synchronization

### How can I request features?

Submit feature requests through:

- **GitHub Issues**: Tag with "enhancement"
- **GitHub Discussions**: Community feature discussions
- **Pull Requests**: Implement and submit the feature

### Is commercial support available?

Pocket Concierge is an open-source project. For commercial deployments:

- **Community support**: GitHub issues and discussions
- **Professional services**: Contact project maintainers
- **Enterprise features**: Custom development available

### What's the project roadmap?

The project focuses on:

1. **Stability**: Bug fixes and reliability improvements
2. **Performance**: Optimization and scaling enhancements
3. **Features**: User-requested functionality
4. **Ecosystem**: Integration with other home network tools

Check the [GitHub project](https://github.com/risadams/Pocket-Concierge) for current roadmap and milestones.
