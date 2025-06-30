# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=v0.1.0-dev
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o pocketconcierge ./cmd/pocketconcierge/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata bind-tools

# Create non-root user
RUN addgroup -g 1001 -S pocketconcierge && \
    adduser -u 1001 -S pocketconcierge -G pocketconcierge

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/pocketconcierge .

# Copy configuration files
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/configs/example.yaml ./configs/

# Change ownership to non-root user
RUN chown -R pocketconcierge:pocketconcierge /app

# Switch to non-root user
USER pocketconcierge

# Expose DNS port
EXPOSE 8053/udp
EXPOSE 8053/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD dig @127.0.0.1 -p 8053 google.com +short >/dev/null 2>&1 || exit 1

# Set default command
ENTRYPOINT ["./pocketconcierge"]
CMD ["/app/config.yaml"]
