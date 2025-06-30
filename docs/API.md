# API Reference

Pocket Concierge DNS server provides a DNS protocol interface for resolving hostnames. This document describes the DNS query interface and behavior.

## DNS Protocol Support

### Supported Query Types

| Query Type | Description | Status |
|------------|-------------|--------|
| A | IPv4 address records | ✅ Supported |
| AAAA | IPv6 address records | ✅ Supported |
| PTR | Reverse DNS lookups | ⚠️ Limited support |
| CNAME | Canonical name records | ❌ Not supported |
| MX | Mail exchange records | ❌ Not supported |
| TXT | Text records | ❌ Not supported |
| SRV | Service records | ❌ Not supported |
| NS | Name server records | ❌ Not supported |

### Supported Classes

| Class | Description | Status |
|-------|-------------|--------|
| IN | Internet class | ✅ Supported |
| CH | Chaos class | ❌ Not supported |
| HS | Hesiod class | ❌ Not supported |

## Query Resolution Process

### 1. Local Host Resolution

When a DNS query is received, Pocket Concierge first checks if the requested hostname matches any configured local hosts.

**Resolution Steps:**

1. **Exact Match**: Check if the query matches a configured hostname exactly
2. **Domain Append**: If no exact match and query is a simple hostname, append the home domain and check again
3. **Case Insensitive**: All hostname matching is case-insensitive

**Example:**

```
Query: "desktop" (A record)
1. Check for exact match: "desktop" → Not found
2. Append home domain: "desktop.home" → Found!
3. Return configured IPv4 addresses
```

### 2. Cache Lookup

If no local host matches, check the DNS cache for recent responses.

**Cache Behavior:**

- **TTL Respect**: Cached entries expire based on their TTL
- **LRU Eviction**: Least recently used entries are evicted when cache is full
- **Thread Safe**: Concurrent access is properly synchronized

### 3. Upstream Resolution

If not in cache, forward the query to upstream DNS servers.

**Upstream Selection:**

- **Sequential**: Try upstream servers in configured order
- **Failover**: Move to next server if current server fails
- **Protocol Support**: UDP, TCP, DNS-over-TLS (DoT), DNS-over-HTTPS (DoH)

## DNS Response Format

### Standard DNS Message Structure

```
DNS Header (12 bytes)
- ID: Query identifier for matching responses
- Flags: QR, Opcode, AA, TC, RD, RA, Z, RCODE
- QDCOUNT: Number of questions
- ANCOUNT: Number of answers
- NSCOUNT: Number of authority records
- ARCOUNT: Number of additional records

Question Section
- QNAME: Domain name being queried
- QTYPE: Query type (A, AAAA, etc.)
- QCLASS: Query class (usually IN)

Answer Section
- NAME: Domain name
- TYPE: Record type
- CLASS: Record class
- TTL: Time to live
- RDLENGTH: Resource data length
- RDATA: Resource data (IP address, etc.)
```

### Response Flags

| Flag | Name | Description | Pocket Concierge Behavior |
|------|------|-------------|---------------------------|
| QR | Query/Response | 0=Query, 1=Response | Always 1 for responses |
| AA | Authoritative Answer | Server is authoritative | Set to 1 for local hosts |
| TC | Truncated | Message was truncated | Set if response too large for UDP |
| RD | Recursion Desired | Client wants recursion | Copied from query |
| RA | Recursion Available | Server supports recursion | Set based on config |

### Response Codes

| Code | Name | Description | When Used |
|------|------|-------------|-----------|
| 0 | NOERROR | No error | Successful resolution |
| 1 | FORMERR | Format error | Malformed query |
| 2 | SERVFAIL | Server failure | Upstream resolution failed |
| 3 | NXDOMAIN | Name does not exist | Hostname not found |
| 4 | NOTIMP | Not implemented | Unsupported query type |
| 5 | REFUSED | Query refused | Recursion disabled |

## Query Examples

### IPv4 Address Query (A Record)

**Query:**

```
; Query for desktop.home A record
;; QUESTION SECTION:
desktop.home.    IN    A

;; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0
```

**Response:**

```
; Response with IPv4 address
;; QUESTION SECTION:
desktop.home.    IN    A

;; ANSWER SECTION:
desktop.home.    300    IN    A    192.168.1.100

;; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
```

### IPv6 Address Query (AAAA Record)

**Query:**

```
; Query for server.home AAAA record
;; QUESTION SECTION:
server.home.    IN    AAAA

;; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0
```

**Response:**

```
; Response with IPv6 address
;; QUESTION SECTION:
server.home.    IN    AAAA

;; ANSWER SECTION:
server.home.    300    IN    AAAA    fe80::1234:5678:90ab:cdef

;; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
```

### Multiple Address Response

**Configuration:**

```yaml
hosts:
  - hostname: "loadbalancer"
    ipv4:
      - "192.168.1.10"
      - "192.168.1.11"
      - "192.168.1.12"
```

**Response:**

```
; Response with multiple IPv4 addresses
;; QUESTION SECTION:
loadbalancer.home.    IN    A

;; ANSWER SECTION:
loadbalancer.home.    300    IN    A    192.168.1.10
loadbalancer.home.    300    IN    A    192.168.1.11
loadbalancer.home.    300    IN    A    192.168.1.12

;; QUERY: 1, ANSWER: 3, AUTHORITY: 0, ADDITIONAL: 0
```

### Non-existent Domain

**Query:**

```
; Query for non-existent host
;; QUESTION SECTION:
nonexistent.home.    IN    A

;; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0
```

**Response:**

```
; NXDOMAIN response
;; QUESTION SECTION:
nonexistent.home.    IN    A

;; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0
;; status: NXDOMAIN
```

## Testing the DNS Server

### Using dig Command

**Basic query:**

```bash
dig @127.0.0.1 -p 8053 desktop.home A
```

**IPv6 query:**

```bash
dig @127.0.0.1 -p 8053 server.home AAAA
```

**Query with full output:**

```bash
dig @127.0.0.1 -p 8053 +all desktop.home A
```

### Using nslookup Command

**Basic query:**

```bash
nslookup desktop.home 127.0.0.1
```

**Specify port:**

```bash
nslookup -port=8053 desktop.home 127.0.0.1
```

### Using the Built-in Test Tool

```bash
go run test-dns.go hostname server:port
```

**Examples:**

```bash
# Test local host resolution
go run test-dns.go desktop.home 127.0.0.1:8053

# Test external resolution
go run test-dns.go google.com 127.0.0.1:8053

# Test IPv6
go run test-dns.go -6 server.home 127.0.0.1:8053
```

## Error Handling

### Common Error Scenarios

**1. Malformed Queries**

- **Cause**: Invalid DNS message format
- **Response**: FORMERR (Format Error)
- **Status Code**: 1

**2. Unsupported Query Types**

- **Cause**: Query for unsupported record type (MX, TXT, etc.)
- **Response**: NOTIMP (Not Implemented)
- **Status Code**: 4

**3. Upstream Server Failure**

- **Cause**: All upstream servers are unreachable
- **Response**: SERVFAIL (Server Failure)
- **Status Code**: 2

**4. Non-existent Domain**

- **Cause**: Hostname not found locally or upstream
- **Response**: NXDOMAIN (Name Does Not Exist)
- **Status Code**: 3

**5. Recursion Disabled**

- **Cause**: Client requests recursion but it's disabled
- **Response**: REFUSED (Query Refused)
- **Status Code**: 5

### Error Response Examples

**Format Error:**

```bash
$ dig @127.0.0.1 -p 8053 malformed..hostname A
; <<>> DiG 9.16.1 <<>> @127.0.0.1 -p 8053 malformed..hostname A
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: FORMERR, id: 12345
```

**Not Implemented:**

```bash
$ dig @127.0.0.1 -p 8053 desktop.home MX
; <<>> DiG 9.16.1 <<>> @127.0.0.1 -p 8053 desktop.home MX
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOTIMP, id: 23456
```

## Performance Characteristics

### Response Times

| Query Type | Typical Response Time |
|------------|----------------------|
| Local host (cached) | < 1ms |
| Local host (uncached) | 1-5ms |
| Upstream (cached) | < 1ms |
| Upstream UDP | 10-50ms |
| Upstream DoT | 20-100ms |
| Upstream DoH | 50-200ms |

### Throughput

- **Concurrent queries**: 1000+ simultaneous queries
- **Queries per second**: 10,000+ QPS for cached responses
- **Cache hit ratio**: 80-95% for typical home networks

### Memory Usage

- **Base memory**: ~10MB
- **Cache memory**: ~100 bytes per cached entry
- **Connection pooling**: ~1KB per upstream connection

## Security Considerations

### Query Validation

- **DNS message format**: Strict RFC compliance
- **Query limits**: Maximum query size limits
- **Rate limiting**: Basic protection against abuse

### Upstream Security

- **Certificate verification**: DoT/DoH certificate validation
- **Connection encryption**: TLS 1.2+ for secure protocols
- **Server validation**: Hostname verification for encrypted connections

### Local Network Security

- **Bind restrictions**: Configurable bind address
- **Access control**: No built-in authentication (assumes trusted network)
- **Logging**: Configurable logging for monitoring

## Limitations

### Protocol Limitations

- **UDP only**: Primary protocol (TCP support for large responses)
- **IPv4/IPv6**: Full support for both protocols
- **EDNS**: Basic EDNS support

### Record Type Limitations

- **A/AAAA only**: Only address records for local hosts
- **No CNAME**: Canonical names not supported
- **No MX/TXT**: Mail/text records not supported

### Deployment Limitations

- **Single instance**: No clustering or replication
- **File-based config**: No dynamic configuration API
- **Home network focus**: Not designed for enterprise environments

## Future API Enhancements

### Planned Features

- **HTTP API**: RESTful configuration interface
- **Metrics endpoint**: Prometheus-compatible metrics
- **Health checks**: HTTP health check endpoints
- **Dynamic updates**: Runtime configuration updates

### Potential Extensions

- **CNAME support**: Canonical name records
- **Wildcard matching**: Pattern-based hostname matching
- **Load balancing**: Advanced load balancing algorithms
- **Monitoring integration**: Enhanced observability features
