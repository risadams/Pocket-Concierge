# Contributing to Pocket Concierge

Thank you for your interest in contributing to Pocket Concierge! 🏨✨

We welcome contributions of all kinds, whether it's bug reports, feature requests, documentation improvements, or code contributions.

## 🤝 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Use the issue templates** when available
3. **Provide clear, detailed information** about the problem
4. **Include system information** (OS, Go version, etc.)

### Suggesting Features

We love new ideas! When suggesting features:

1. **Check if it aligns** with the project's goals
2. **Describe the use case** clearly
3. **Consider backwards compatibility**
4. **Discuss implementation approaches**

### Contributing Code

#### Getting Started

1. **Fork the repository**
2. **Clone your fork** locally
3. **Create a feature branch** from `main`
4. **Set up your development environment**

```bash
# Clone your fork
git clone https://github.com/your-username/Pocket-Concierge.git
cd Pocket-Concierge

# Add upstream remote
git remote add upstream https://github.com/risadams/Pocket-Concierge.git

# Create a feature branch
git checkout -b feature/your-feature-name
```

#### Development Environment

**Prerequisites:**

- Go 1.24 or later
- Make (for build automation)
- Git

**Setup:**

```bash
# Install dependencies
go mod download

# Run tests to ensure everything works
make test

# Build the project
make build
```

#### Making Changes

1. **Write clean, idiomatic Go code**
2. **Follow existing code style**
3. **Add tests for new functionality**
4. **Update documentation** as needed
5. **Keep commits focused and atomic**

#### Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (run `make fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

#### Testing

- **Write tests** for all new functionality
- **Maintain or improve** test coverage
- **Run the full test suite** before submitting

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make benchmark
```

#### Documentation

- Update README.md if needed
- Add or update code comments
- Update configuration examples
- Consider adding examples for new features

### Submitting Pull Requests

1. **Ensure your branch is up to date** with main
2. **Run all tests and linting**
3. **Write a clear PR description**
4. **Reference any related issues**
5. **Be responsive to feedback**

```bash
# Before submitting, ensure everything passes
make fmt
make lint
make test
make build
```

#### Pull Request Guidelines

- **Use a clear, descriptive title**
- **Describe what changed and why**
- **Include screenshots** for UI changes
- **Reference issues** using "Fixes #123" or "Closes #123"
- **Keep PRs focused** - one feature/fix per PR
- **Update CHANGELOG.md** if applicable

## 🏗️ Project Structure

```
├── cmd/                    # Main applications
│   ├── pocketconcierge/   # Main DNS server
│   ├── benchmark/         # Benchmarking tool
│   └── loadtest/          # Load testing tool
├── internal/              # Private application code
│   ├── config/           # Configuration handling
│   ├── dns/              # DNS logic and handlers
│   └── server/           # Server implementation
├── configs/              # Configuration examples
├── test/                 # Integration tests
├── build/                # Build artifacts
└── docs/                 # Additional documentation
```

## 🧪 Testing Guidelines

### Unit Tests

- Test all public functions
- Use table-driven tests for multiple inputs
- Mock external dependencies
- Test error conditions

### Integration Tests

- Test complete workflows
- Use real network connections when appropriate
- Test configuration loading
- Verify DNS resolution chains

### Benchmarks

- Benchmark performance-critical code
- Include memory allocation benchmarks
- Compare before and after performance

## 📋 Code Review Process

1. **Automated checks** run on all PRs
2. **Maintainer review** for code quality and design
3. **Testing** on multiple platforms if needed
4. **Documentation review** for user-facing changes
5. **Merge** after approval and passing checks

## 🚀 Release Process

1. Version bumping follows [Semantic Versioning](https://semver.org/)
2. Releases are tagged and include release notes
3. Binary releases are built automatically
4. Docker images are updated for releases

## 🎯 Development Best Practices

### Security

- Never commit secrets or credentials
- Validate all inputs
- Use secure defaults
- Follow security best practices for DNS

### Performance

- Profile code for performance bottlenecks
- Optimize memory allocations
- Use efficient data structures
- Consider concurrency implications

### Compatibility

- Maintain backwards compatibility when possible
- Document breaking changes clearly
- Support multiple Go versions when feasible
- Test on different operating systems

## 📚 Resources

- [Go Documentation](https://golang.org/doc/)
- [DNS RFC 1035](https://tools.ietf.org/html/rfc1035)
- [DNS-over-HTTPS RFC 8484](https://tools.ietf.org/html/rfc8484)
- [DNS-over-TLS RFC 7858](https://tools.ietf.org/html/rfc7858)

## 🎉 Recognition

Contributors are recognized in:

- CHANGELOG.md for their contributions
- GitHub contributor stats
- Release notes for significant contributions

## ❓ Questions?

- Open a [Discussion](https://github.com/risadams/Pocket-Concierge/discussions)
- Check existing [Issues](https://github.com/risadams/Pocket-Concierge/issues)
- Review the [README](README.md)

Thank you for helping make Pocket Concierge better! 🙏
