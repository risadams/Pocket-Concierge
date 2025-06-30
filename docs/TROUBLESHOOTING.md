# Troubleshooting Guide

This guide helps diagnose and resolve common issues with Pocket Concierge DNS server.

## Quick Diagnostics

### Basic Health Check

First, verify that Pocket Concierge is running and responding:

```bash
# Check if process is running
ps aux | grep pocketconcierge

# Test basic DNS response
dig @127.0.0.1 -p 8053 google.com

# Test local hostname resolution
dig @127.0.0.1 -p 8053 desktop.home
```

### Service Status Check

```bash
# For systemd systems
sudo systemctl status pocketconcierge

# Check recent logs
sudo journalctl -u pocketconcierge --lines=50

# For Docker deployments
docker ps | grep pocketconcierge
docker logs pocketconcierge
```

## Common Issues and Solutions

### 1. Server Won't Start

#### Issue: Permission denied binding to port 53

**Error message:**

```
❌ Failed to start server: listen udp :53: bind: permission denied
```

**Cause:** Non-root user trying to bind to privileged port (< 1024)

**Solutions:**

**Option A: Run with sudo (not recommended for production)**

```bash
sudo ./pocketconcierge
```

**Option B: Use setcap to grant binding permission (Linux)**

```bash
sudo setcap CAP_NET_BIND_SERVICE=+eip ./pocketconcierge
./pocketconcierge
```

**Option C: Use non-privileged port**

```yaml
server:
  port: 8053  # Use port > 1024
```

**Option D: Use systemd service with User directive**

```ini
[Service]
User=root  # Or configure capabilities
```

#### Issue: Config file not found

**Error message:**

```
⚠️ Config loading failed: open config.yaml: no such file or directory
```

**Solutions:**

```bash
# Verify file exists
ls -la config.yaml

# Check current directory
pwd

# Specify full path
./pocketconcierge /full/path/to/config.yaml

# Copy example config
cp configs/example.yaml config.yaml
```

#### Issue: Invalid configuration

**Error message:**

```
⚠️ Config loading failed: yaml: unmarshal errors
```

**Solutions:**

```bash
# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"

# Check for tabs (YAML requires spaces)
cat -A config.yaml | grep -P '\t'

# Use example config as reference
diff config.yaml configs/example.yaml
```

### 2. DNS Resolution Issues

#### Issue: Local hostnames not resolving

**Symptoms:**

- External domains resolve fine
- Configured local hostnames return NXDOMAIN

**Debugging steps:**

1. **Verify configuration:**

```yaml
# Check hosts section in config.yaml
hosts:
  - hostname: "desktop"
    ipv4:
      - "192.168.1.100"

# Verify home domain setting
home_dns_domain: "home"
```

2. **Test with debug logging:**

```yaml
log_level: "debug"
```

3. **Check exact hostname format:**

```bash
# Try both with and without domain suffix
dig @127.0.0.1 -p 8053 desktop
dig @127.0.0.1 -p 8053 desktop.home
```

4. **Verify case sensitivity:**

```bash
# DNS is case-insensitive, these should work the same
dig @127.0.0.1 -p 8053 DESKTOP.HOME
dig @127.0.0.1 -p 8053 desktop.home
```

#### Issue: External domains not resolving

**Symptoms:**

- Local hostnames work
- External domains timeout or fail

**Debugging steps:**

1. **Check upstream configuration:**

```yaml
upstream:
  - name: "Test Server"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53
```

2. **Test upstream connectivity:**

```bash
# Test upstream server directly
dig @1.1.1.1 google.com

# Test network connectivity
ping 1.1.1.1
telnet 1.1.1.1 53
```

3. **Check firewall rules:**

```bash
# Linux: Check iptables
sudo iptables -L OUTPUT | grep 53

# Test with firewall disabled temporarily
sudo ufw disable  # Ubuntu
# Test DNS resolution
sudo ufw enable   # Re-enable
```

4. **Verify DNS over HTTPS/TLS:**

```bash
# For DoH, test HTTPS connectivity
curl -v https://cloudflare-dns.com/dns-query

# For DoT, test TLS connectivity
openssl s_client -connect 1.1.1.1:853 -servername cloudflare-dns.com
```

#### Issue: Intermittent resolution failures

**Symptoms:**

- DNS sometimes works, sometimes fails
- Timeout errors

**Possible causes:**

1. **Network connectivity issues:**

```bash
# Monitor network connectivity
ping -c 10 1.1.1.1

# Check for packet loss
mtr 1.1.1.1
```

2. **Upstream server problems:**

```yaml
# Add multiple upstream servers for redundancy
upstream:
  - name: "Cloudflare Primary"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53
  
  - name: "Google Fallback"
    address: "8.8.8.8"
    protocol: "udp"
    port: 53
```

3. **Cache issues:**

```yaml
# Try disabling cache temporarily
dns:
  cache_size: 0  # Disable caching
```

### 3. Performance Issues

#### Issue: Slow DNS responses

**Symptoms:**

- DNS queries take several seconds
- Timeouts on some queries

**Debugging steps:**

1. **Measure response times:**

```bash
# Time DNS queries
time dig @127.0.0.1 -p 8053 google.com

# Use benchmark tool
go run cmd/benchmark/main.go
```

2. **Check cache configuration:**

```yaml
dns:
  cache_size: 10000  # Increase cache size
  ttl: 300          # Reasonable TTL
```

3. **Optimize upstream servers:**

```yaml
upstream:
  # Put fastest servers first
  - name: "Local ISP DNS"
    address: "192.168.1.1"  # Router DNS
    protocol: "udp"
    port: 53
  
  - name: "Cloudflare"
    address: "1.1.1.1"
    protocol: "udp"
    port: 53
```

4. **Monitor system resources:**

```bash
# Check CPU usage
top -p $(pgrep pocketconcierge)

# Check memory usage
ps -o pid,vsz,rss,comm -p $(pgrep pocketconcierge)

# Check network connections
netstat -tulpn | grep pocketconcierge
```

#### Issue: High memory usage

**Symptoms:**

- Process consuming excessive RAM
- System becoming slow

**Solutions:**

1. **Reduce cache size:**

```yaml
dns:
  cache_size: 1000  # Reduce from default 10000
```

2. **Monitor cache statistics:**

```bash
# Enable debug logging to see cache hits/misses
log_level: "debug"
```

3. **Restart service periodically:**

```bash
# Add to cron for daily restart
0 3 * * * systemctl restart pocketconcierge
```

### 4. Network Configuration Issues

#### Issue: Clients not using Pocket Concierge

**Symptoms:**

- DNS server running but clients still use default DNS
- Local hostnames not resolving on client devices

**Solutions:**

1. **Check client DNS configuration:**

**Windows:**

```cmd
ipconfig /all | findstr "DNS Servers"
```

**macOS:**

```bash
scutil --dns | grep nameserver
```

**Linux:**

```bash
cat /etc/resolv.conf
systemd-resolve --status
```

2. **Verify DHCP configuration:**

```bash
# Check DHCP lease
cat /var/lib/dhcp/dhclient.leases  # Linux
# Look for "option domain-name-servers"
```

3. **Test from client machine:**

```bash
# Test DNS resolution from client
nslookup desktop.home <pocket-concierge-ip>
```

4. **Force DNS configuration:**

**Windows:**

```cmd
# Set DNS manually
netsh interface ip set dns "Local Area Connection" static 192.168.1.100
```

**macOS:**

```bash
# Set DNS via NetworkSetup
sudo networksetup -setdnsservers "Ethernet" 192.168.1.100
```

#### Issue: DNS conflicts with existing server

**Symptoms:**

- Both servers responding to queries
- Inconsistent resolution results

**Solutions:**

1. **Check for other DNS services:**

```bash
# Linux: Check for other services on port 53
sudo netstat -tulpn | grep :53
sudo lsof -i :53

# Common conflicting services
sudo systemctl status systemd-resolved
sudo systemctl status dnsmasq
sudo systemctl status bind9
```

2. **Disable conflicting services:**

```bash
# Disable systemd-resolved
sudo systemctl disable systemd-resolved
sudo systemctl stop systemd-resolved

# Or configure it to not bind to port 53
sudo mkdir -p /etc/systemd/resolved.conf.d
echo "[Resolve]
DNSStubListener=no" | sudo tee /etc/systemd/resolved.conf.d/no-stub.conf
```

3. **Use different port temporarily:**

```yaml
server:
  port: 8053  # Use non-conflicting port
```

### 5. Certificate and Encryption Issues

#### Issue: DoT/DoH upstream failures

**Error message:**

```
Failed to connect to upstream: x509: certificate verify failed
```

**Solutions:**

1. **Check certificate verification:**

```yaml
upstream:
  - name: "Cloudflare DoT"
    address: "1.1.1.1"
    protocol: "tls"
    port: 853
    verify: true  # Try setting to false temporarily
```

2. **Test TLS connectivity:**

```bash
# Test DoT connection
openssl s_client -connect 1.1.1.1:853 -servername cloudflare-dns.com

# Test DoH connection
curl -v https://cloudflare-dns.com/dns-query
```

3. **Check system time:**

```bash
# Verify system time is correct (affects certificate validation)
date
ntpdate -q pool.ntp.org
```

4. **Update CA certificates:**

```bash
# Linux
sudo apt update && sudo apt install ca-certificates

# macOS
# Usually updated with system updates
```

## Advanced Debugging

### Enable Debug Logging

```yaml
log_level: "debug"
```

This provides detailed information about:

- Incoming DNS queries
- Cache hits and misses
- Upstream server communications
- Error conditions

### Packet Capture

Use packet capture to debug network-level issues:

```bash
# Capture DNS traffic
sudo tcpdump -i any port 53 -v

# Capture traffic to/from specific host
sudo tcpdump -i any host 1.1.1.1 -v

# Save capture for analysis
sudo tcpdump -i any port 53 -w dns-capture.pcap
```

### DNS Query Testing

Test different query types and sources:

```bash
# Test different record types
dig @127.0.0.1 -p 8053 desktop.home A
dig @127.0.0.1 -p 8053 desktop.home AAAA

# Test with different tools
nslookup desktop.home 127.0.0.1
host desktop.home 127.0.0.1

# Test from remote machine
dig @<pocket-concierge-ip> -p 8053 desktop.home
```

### Performance Profiling

```bash
# Run load test
go run cmd/loadtest/main.go

# Monitor with system tools
htop
iotop
nethogs
```

## Error Reference

### Common Error Codes

| Error | Meaning | Common Causes |
|-------|---------|---------------|
| NXDOMAIN | Name does not exist | Hostname not configured, typo in hostname |
| SERVFAIL | Server failure | Upstream DNS failure, network issues |
| FORMERR | Format error | Malformed DNS query |
| REFUSED | Query refused | Recursion disabled for external queries |
| TIMEOUT | Query timeout | Network connectivity issues |

### Log Message Reference

**Normal operation:**

```
✅ Ready to serve your home network!
📋 Server: 127.0.0.1:8053
```

**Configuration issues:**

```
⚠️ Config loading failed: <error details>
```

**Network issues:**

```
❌ Failed to start server: <error details>
❌ Error stopping server: <error details>
```

## Getting Help

### Collect Debug Information

Before seeking help, collect this information:

1. **Configuration file:**

```bash
cat config.yaml
```

2. **Version information:**

```bash
./pocketconcierge --version  # If implemented
go version
```

3. **System information:**

```bash
uname -a
cat /etc/os-release
```

4. **Log output with debug enabled:**

```bash
# Run with debug logging
./pocketconcierge 2>&1 | tee debug.log
```

5. **Network configuration:**

```bash
ip addr show
cat /etc/resolv.conf
```

### Test Cases

Provide results of these tests:

```bash
# Basic connectivity
ping 1.1.1.1

# Direct upstream test
dig @1.1.1.1 google.com

# Local server test
dig @127.0.0.1 -p 8053 google.com

# Local hostname test
dig @127.0.0.1 -p 8053 desktop.home
```

### Support Channels

- GitHub Issues: [Report bugs and feature requests](https://github.com/risadams/Pocket-Concierge/issues)
- Discussions: [Community support and questions](https://github.com/risadams/Pocket-Concierge/discussions)
- Documentation: [Check latest documentation](https://github.com/risadams/Pocket-Concierge/docs)

### Contributing Fixes

If you find and fix an issue:

1. Create a test case that reproduces the issue
2. Implement the fix
3. Verify the fix resolves the issue
4. Submit a pull request with test case and fix

This helps improve Pocket Concierge for everyone!
