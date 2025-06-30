# Docker Deployment Guide for PocketConcierge

This guide covers how to deploy PocketConcierge using Docker and Docker Compose.

## Quick Start

### Using Make Commands

The simplest way to get started is using the included Makefile:

```bash
# Build and run the container
make docker-all

# Or step by step:
make docker-build    # Build the Docker image
make docker-run      # Run the container interactively
```

### Using Docker Directly

```bash
# Build the image
docker build -t pocketconcierge:latest .

# Run the container
docker run --rm -it \
  -p 8053:8053/udp \
  -p 8053:8053/tcp \
  pocketconcierge:latest
```

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

## Available Make Commands

### Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image |
| `make docker-build-no-cache` | Build Docker image without cache |
| `make docker-run` | Run container interactively |
| `make docker-run-daemon` | Run container as daemon |
| `make docker-run-custom CONFIG=path/to/config.yaml` | Run with custom config |
| `make docker-stop` | Stop running container |
| `make docker-logs` | Show container logs |
| `make docker-shell` | Open shell in container |
| `make docker-push DOCKER_REGISTRY=your-registry.com/username` | Push to registry |
| `make docker-clean` | Remove containers and images |
| `make docker-all` | Build and run container |

### Docker Compose Commands

| Command | Description |
|---------|-------------|
| `make compose-up` | Start services |
| `make compose-up-build` | Build and start services |
| `make compose-down` | Stop services |
| `make compose-logs` | Show logs |
| `make compose-restart` | Restart services |
| `make compose-clean` | Clean resources |
| `make compose-dev` | Start development environment |

### Multi-Architecture Commands

| Command | Description |
|---------|-------------|
| `make docker-buildx-setup` | Set up buildx for multi-arch |
| `make docker-buildx-multiarch` | Build for multiple architectures |
| `make docker-buildx-local` | Build multi-arch locally |

## Configuration

### Environment Variables

You can customize the Docker deployment using environment variables:

```bash
# Set custom version
export VERSION=v1.0.0

# Set custom registry
export DOCKER_REGISTRY=docker.io/username

# Build with custom settings
make docker-build
```

### Custom Configuration

To use a custom configuration file:

```bash
# Using make command
make docker-run-custom CONFIG=my-config.yaml

# Using Docker directly
docker run --rm -it \
  -p 8053:8053/udp \
  -p 8053:8053/tcp \
  -v $(pwd)/my-config.yaml:/app/config.yaml:ro \
  pocketconcierge:latest
```

### Persistent Data

To persist logs or other data:

```bash
docker run -d \
  -p 8053:8053/udp \
  -p 8053:8053/tcp \
  -v $(pwd)/logs:/app/logs \
  --name pocketconcierge \
  pocketconcierge:latest
```

## Deployment Scripts

### Linux/macOS (deploy.sh)

```bash
# Make executable
chmod +x deploy.sh

# Deploy with defaults
./deploy.sh deploy

# Deploy on custom port as daemon
./deploy.sh deploy -p 53 -d

# Deploy with custom config
./deploy.sh deploy -c custom.yaml

# Check status
./deploy.sh status

# View logs
./deploy.sh logs

# Stop service
./deploy.sh stop

# Clean up
./deploy.sh clean
```

### Windows PowerShell (deploy.ps1)

```powershell
# Deploy with defaults
.\deploy.ps1 deploy

# Deploy on custom port as daemon
.\deploy.ps1 deploy -Port 53 -Daemon

# Deploy with custom config
.\deploy.ps1 deploy -Config custom.yaml

# Check status
.\deploy.ps1 status

# View logs
.\deploy.ps1 logs

# Stop service
.\deploy.ps1 stop

# Clean up
.\deploy.ps1 clean
```

## Production Deployment

### Using Docker Compose for Production

1. Edit `docker-compose.yml` to suit your needs:

   ```yaml
   version: '3.8'
   services:
     pocketconcierge:
       image: pocketconcierge:latest
       container_name: pocketconcierge
       restart: unless-stopped
       ports:
         - "53:8053/udp"  # Use port 53 for production
         - "53:8053/tcp"
       volumes:
         - ./config.yaml:/app/config.yaml:ro
         - ./logs:/app/logs
   ```

2. Deploy:

   ```bash
   docker-compose up -d
   ```

### Security Considerations

The Docker image includes several security features:

- **Non-root user**: Runs as user `pocketconcierge` (UID 1001)
- **Read-only filesystem**: Container filesystem is read-only
- **Dropped capabilities**: Minimal required capabilities only
- **Security options**: No new privileges allowed

### Multi-Architecture Support

Build and push multi-architecture images:

```bash
# Set up buildx
make docker-buildx-setup

# Build for multiple architectures and push
make docker-buildx-multiarch DOCKER_REGISTRY=your-registry.com/username
```

Supported architectures:

- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon, Raspberry Pi 4+)
- `linux/arm/v7` (ARM 32-bit, Raspberry Pi 3)

## Troubleshooting

### Common Issues

1. **Permission denied on port 53**:

   ```bash
   # Use unprivileged port or run as root
   sudo docker run -p 53:8053/udp -p 53:8053/tcp pocketconcierge:latest
   ```

2. **Container exits immediately**:

   ```bash
   # Check logs
   docker logs pocketconcierge
   
   # Run interactively to debug
   docker run --rm -it pocketconcierge:latest /bin/sh
   ```

3. **Config file not found**:

   ```bash
   # Ensure config file exists and path is correct
   ls -la config.yaml
   
   # Check mount path
   docker run --rm -it -v $(pwd)/config.yaml:/app/config.yaml:ro pocketconcierge:latest
   ```

### Health Checks

The container includes health checks that verify DNS functionality:

```bash
# Check health status
docker inspect pocketconcierge | grep -A 10 Health

# Manual health check
docker exec pocketconcierge nslookup google.com 127.0.0.1:8053
```

### Performance Monitoring

Monitor container resource usage:

```bash
# Real-time stats
docker stats pocketconcierge

# Container processes
docker exec pocketconcierge ps aux

# Memory usage
docker exec pocketconcierge cat /proc/meminfo
```

## Advanced Configuration

### Custom Dockerfile

You can customize the Dockerfile for your specific needs:

```dockerfile
FROM pocketconcierge:latest

# Add custom configurations
COPY custom-config.yaml /app/config.yaml

# Add custom scripts
COPY scripts/ /app/scripts/

# Set custom environment variables
ENV LOG_LEVEL=debug
ENV DNS_PORT=8053
```

### Integration with Monitoring

Add monitoring to your Docker Compose setup:

```yaml
version: '3.8'
services:
  pocketconcierge:
    # ... existing config ...
  
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

This comprehensive setup provides a robust foundation for deploying PocketConcierge in various environments using Docker.
