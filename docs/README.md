# Pocket Concierge Documentation

Welcome to the comprehensive documentation for Pocket Concierge DNS server. This documentation covers everything you need to know about installing, configuring, deploying, and troubleshooting your home network DNS solution.

## Quick Start

New to Pocket Concierge? Start here:

1. **[README](../README.md)** - Project overview and quick installation guide
2. **[Configuration Guide](CONFIGURATION.md)** - Learn how to configure your DNS server
3. **[Deployment Guide](DEPLOYMENT.md)** - Deploy in your home network

## Documentation Index

### Core Documentation

| Document | Description | Audience |
|----------|-------------|----------|
| **[Configuration Reference](CONFIGURATION.md)** | Complete configuration options and examples | All users |
| **[API Reference](API.md)** | DNS protocol interface and behavior | Advanced users, developers |
| **[Architecture Overview](ARCHITECTURE.md)** | System design and internal components | Developers, system architects |

### Deployment and Operations

| Document | Description | Audience |
|----------|-------------|----------|
| **[Deployment Guide](DEPLOYMENT.md)** | Installation scenarios and system integration | System administrators |
| **[Troubleshooting Guide](TROUBLESHOOTING.md)** | Common issues and debugging techniques | All users |
| **[FAQ](FAQ.md)** | Frequently asked questions and answers | All users |

### Project Information

| Document | Description |
|----------|-------------|
| **[Contributing Guide](../CONTRIBUTING.md)** | How to contribute to the project |
| **[Code of Conduct](../CODE_OF_CONDUCT.md)** | Community guidelines |
| **[Security Policy](../SECURITY.md)** | Security practices and reporting |
| **[License](../LICENSE)** | Project license information |

## Getting Started Workflow

### For Home Users

1. **Learn the basics**: Read the [FAQ](FAQ.md) to understand what Pocket Concierge does
2. **Quick setup**: Follow the [README](../README.md) quick start guide
3. **Configure**: Use the [Configuration Guide](CONFIGURATION.md) to set up your hosts
4. **Deploy**: Follow the [Deployment Guide](DEPLOYMENT.md) for your specific scenario
5. **Troubleshoot**: Use the [Troubleshooting Guide](TROUBLESHOOTING.md) if you encounter issues

### For Developers

1. **Understand the architecture**: Read the [Architecture Overview](ARCHITECTURE.md)
2. **Study the API**: Review the [API Reference](API.md)
3. **Set up development**: Follow the build instructions in [README](../README.md)
4. **Contribute**: See the [Contributing Guide](../CONTRIBUTING.md)

### For System Administrators

1. **Review deployment options**: Study the [Deployment Guide](DEPLOYMENT.md)
2. **Plan configuration**: Use the [Configuration Reference](CONFIGURATION.md)
3. **Understand troubleshooting**: Familiarize yourself with the [Troubleshooting Guide](TROUBLESHOOTING.md)
4. **Monitor operations**: Review monitoring and maintenance sections

## Common Use Cases

### Personal Computer DNS Override

**Goal**: Use custom hostnames on your personal computer

**Documents to read**:

- [Configuration Guide](CONFIGURATION.md) - Basic host configuration
- [Deployment Guide](DEPLOYMENT.md) - Personal computer scenario
- [FAQ](FAQ.md) - Client configuration questions

### Home Network DNS Server

**Goal**: Provide DNS services for your entire home network

**Documents to read**:

- [Deployment Guide](DEPLOYMENT.md) - Router integration and Raspberry Pi setup
- [Configuration Guide](CONFIGURATION.md) - Multiple upstream servers and caching
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Network configuration issues

### Development and Testing

**Goal**: Set up a development environment or test configuration changes

**Documents to read**:

- [Architecture Overview](ARCHITECTURE.md) - Understanding the system
- [API Reference](API.md) - DNS protocol details
- [Configuration Guide](CONFIGURATION.md) - Development environment settings

## Key Concepts

### DNS Resolution Flow

Understanding how Pocket Concierge resolves DNS queries:

1. **Local hosts check**: Match against configured hostnames
2. **Cache lookup**: Check for cached responses
3. **Upstream resolution**: Forward to configured DNS servers
4. **Response caching**: Store successful responses for future use

### Configuration Hierarchy

Configuration priority and structure:

1. **Server settings**: Network binding and protocol options
2. **DNS behavior**: Caching, recursion, and TTL settings
3. **Upstream servers**: External DNS servers and protocols
4. **Local hosts**: Custom hostname-to-IP mappings

### Deployment Patterns

Common deployment scenarios:

- **Single user**: DNS override for one computer
- **Home network**: DNS server for all home devices
- **Container**: Docker deployment for portability
- **Service**: System service for reliability

## Feature Matrix

| Feature | Status | Documentation |
|---------|--------|---------------|
| Local hostname resolution | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| DNS caching | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| DNS-over-HTTPS (DoH) | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| DNS-over-TLS (DoT) | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| IPv4 and IPv6 support | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| Multiple upstream servers | ✅ Available | [Configuration Guide](CONFIGURATION.md) |
| Systemd service | ✅ Available | [Deployment Guide](DEPLOYMENT.md) |
| Docker support | ✅ Available | [Deployment Guide](DEPLOYMENT.md) |
| Benchmarking tools | ✅ Available | [API Reference](API.md) |
| Web interface | ❌ Not available | Planned for future release |
| Domain blocking | ❌ Not available | Planned for future release |
| Wildcard hostnames | ❌ Not available | Planned for future release |

## Support and Community

### Getting Help

1. **Documentation first**: Check these docs for answers
2. **FAQ**: Review common questions and solutions
3. **GitHub Issues**: Report bugs or request features
4. **GitHub Discussions**: Ask questions and share experiences

### Contributing

We welcome contributions in many forms:

- **Documentation**: Improve or expand these docs
- **Bug reports**: Help us identify and fix issues
- **Feature requests**: Suggest new functionality
- **Code contributions**: Submit pull requests
- **Testing**: Test on different platforms and scenarios

See the [Contributing Guide](../CONTRIBUTING.md) for details.

### Community Guidelines

Please follow our [Code of Conduct](../CODE_OF_CONDUCT.md) when participating in the community. We strive to maintain a welcoming and inclusive environment for all contributors.

## Documentation Maintenance

This documentation is actively maintained and updated with each release. If you find errors or have suggestions for improvement:

1. **File an issue**: Report documentation bugs
2. **Submit a PR**: Fix errors or add improvements
3. **Start a discussion**: Suggest major documentation changes

### Documentation Standards

Our documentation follows these principles:

- **Clarity**: Clear, concise explanations
- **Completeness**: Comprehensive coverage of features
- **Examples**: Practical examples for all concepts
- **Accessibility**: Easy to navigate and search
- **Accuracy**: Kept up-to-date with software changes

## Version Information

This documentation corresponds to Pocket Concierge v0.1.0. For the latest documentation, visit the [GitHub repository](https://github.com/risadams/Pocket-Concierge).

### Changelog

Major documentation updates are tracked in the project changelog. See the [releases page](https://github.com/risadams/Pocket-Concierge/releases) for version-specific changes.

---

**Need help?** Start with the [FAQ](FAQ.md) or [Troubleshooting Guide](TROUBLESHOOTING.md). For additional support, visit our [GitHub Discussions](https://github.com/risadams/Pocket-Concierge/discussions).
