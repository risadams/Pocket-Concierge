# Architecture Overview

## System Architecture

Pocket Concierge is designed as a lightweight, high-performance DNS server optimized for home networks. The architecture follows a modular design with clear separation of concerns.

```text
┌─────────────────────────────────────────────────────────┐
│                    Client Devices                      │
│           (laptops, phones, IoT devices)               │
└─────────────────────┬───────────────────────────────────┘
                      │ DNS Queries (UDP/TCP Port 8053)
                      │
┌─────────────────────▼───────────────────────────────────┐
│                Pocket Concierge                        │
│  ┌─────────────────────────────────────────────────────┐│
│  │              DNS Handler                            ││
│  │  • Query parsing and validation                     ││
│  │  • Request routing logic                            ││
│  │  • Response formatting                              ││
│  └─────────────────────┬───────────────────────────────┘│
│                        │                                │
│  ┌─────────────────────▼───────────────────────────────┐│
│  │             Local Resolver                          ││
│  │  • Home domain resolution (.home)                   ││
│  │  • Static host mapping                              ││
│  │  • IPv4/IPv6 address resolution                     ││
│  └─────────────────────┬───────────────────────────────┘│
│                        │                                │
│  ┌─────────────────────▼───────────────────────────────┐│
│  │               DNS Cache                             ││
│  │  • In-memory LRU cache                              ││
│  │  • Configurable size and TTL                        ││
│  │  • Thread-safe operations                           ││
│  └─────────────────────┬───────────────────────────────┘│
│                        │                                │
│  ┌─────────────────────▼───────────────────────────────┐│
│  │            Upstream Resolver                        ││
│  │  • Multiple upstream servers                        ││
│  │  • Protocol support (UDP/TCP/DoH/DoT)               ││
│  │  • Automatic failover                               ││
│  └─────────────────────┬───────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
                        │
                        │ Secure DNS Queries
                        │
┌─────────────────────▼───────────────────────────────────┐
│              Upstream DNS Servers                      │
│  • DNS-over-HTTPS (DoH) - Port 443                     │
│  • DNS-over-TLS (DoT) - Port 853                       │
│  • Traditional DNS (UDP/TCP) - Port 53                 │
└─────────────────────────────────────────────────────────┘
```

## Core Components

### 1. DNS Handler (`internal/dns/handler.go`)

- **Purpose**: Entry point for all DNS queries
- **Responsibilities**:
  - Parse incoming DNS queries
  - Route queries to appropriate resolvers
  - Format and send DNS responses
  - Handle protocol-specific logic (UDP/TCP)

### 2. Local Resolver (`internal/dns/resolver.go`)

- **Purpose**: Resolve local hostnames and custom mappings
- **Responsibilities**:
  - Match queries against configured hosts
  - Append home domain suffix to simple hostnames
  - Return IPv4/IPv6 addresses for local hosts
  - Handle wildcard and pattern matching

### 3. DNS Cache (`internal/dns/cache.go`)

- **Purpose**: Cache DNS responses to improve performance
- **Responsibilities**:
  - Store successful DNS responses
  - Implement LRU eviction policy
  - Respect TTL values
  - Thread-safe operations for concurrent access

### 4. Secure Client (`internal/dns/secureclient.go`)

- **Purpose**: Handle secure upstream DNS protocols
- **Responsibilities**:
  - DNS-over-HTTPS (DoH) implementation
  - DNS-over-TLS (DoT) implementation
  - Certificate verification
  - Connection pooling and reuse

### 5. Configuration Manager (`internal/config/config.go`)

- **Purpose**: Load and manage application configuration
- **Responsibilities**:
  - Parse YAML configuration files
  - Validate configuration parameters
  - Provide default values
  - Support runtime configuration updates

### 6. Server (`internal/server/server.go`)

- **Purpose**: Main server orchestration
- **Responsibilities**:
  - Initialize all components
  - Manage server lifecycle
  - Handle graceful shutdown
  - Coordinate between components

## Data Flow

### Query Processing Flow

1. **Client Query**: A client device sends a DNS query to Pocket Concierge
2. **Handler Reception**: DNS handler receives and parses the query
3. **Local Resolution Check**: System checks if query matches local hosts
4. **Cache Lookup**: If not local, check if response is cached
5. **Upstream Query**: If cache miss, forward to upstream DNS servers
6. **Response Caching**: Cache successful upstream responses
7. **Client Response**: Send formatted response back to client

### Configuration Flow

1. **Config Loading**: Load YAML configuration at startup
2. **Validation**: Validate all configuration parameters
3. **Component Initialization**: Initialize components with config
4. **Runtime Updates**: Support hot-reloading of certain config changes

## Security Considerations

### DNS Security

- **DoH/DoT Support**: Encrypted DNS queries to upstream servers
- **Certificate Verification**: Validate upstream server certificates
- **Query Validation**: Sanitize and validate all incoming queries
- **Rate Limiting**: Prevent DNS amplification attacks

### Network Security

- **Bind Address**: Configurable bind address for security
- **Local Network Only**: Designed for trusted home networks
- **No Authentication**: Assumes trusted local environment

## Performance Optimizations

### Caching Strategy

- **LRU Cache**: Efficient memory usage with LRU eviction
- **TTL Respect**: Honor upstream TTL values
- **Concurrent Access**: Thread-safe cache operations

### Connection Management

- **Connection Pooling**: Reuse connections to upstream servers
- **Timeout Management**: Appropriate timeouts for all operations
- **Graceful Degradation**: Fallback to alternative protocols

### Memory Management

- **Bounded Cache**: Configurable cache size limits
- **Efficient Data Structures**: Optimized for DNS workloads
- **Garbage Collection**: Minimal GC pressure

## Monitoring and Observability

### Built-in Tools

- **Benchmark Tool**: Performance testing utilities
- **Load Test Tool**: Stress testing capabilities
- **Health Checks**: Basic server health monitoring

### Logging

- **Structured Logging**: Consistent log format
- **Log Levels**: Configurable verbosity
- **Performance Metrics**: Request timing and success rates

## Extensibility

The modular architecture allows for easy extension:

### Adding New Protocols

- Implement new resolver in `internal/dns/`
- Register with handler routing logic
- Update configuration schema

### Custom Resolution Logic

- Extend local resolver with new matching rules
- Add new host configuration options
- Implement custom query processing

### Enhanced Monitoring

- Add metrics collection interfaces
- Implement health check endpoints
- Export monitoring data

## Dependencies

### Core Dependencies

- **github.com/miekg/dns**: DNS protocol implementation
- **gopkg.in/yaml.v3**: YAML configuration parsing

### Standard Library Usage

- **net**: Network operations
- **crypto/tls**: TLS connections
- **sync**: Concurrency primitives
- **context**: Request context management
