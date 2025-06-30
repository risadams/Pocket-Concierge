#!/bin/bash

# PocketConcierge Docker Deployment Script
# This script helps deploy PocketConcierge using Docker

set -e

PROJECT_NAME="pocketconcierge"
DEFAULT_PORT="8053"
DEFAULT_CONFIG="config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_help() {
    echo "PocketConcierge Docker Deployment Script"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy    Deploy PocketConcierge container"
    echo "  stop      Stop PocketConcierge container"
    echo "  restart   Restart PocketConcierge container"
    echo "  logs      Show container logs"
    echo "  status    Show container status"
    echo "  clean     Remove container and image"
    echo "  help      Show this help message"
    echo ""
    echo "Options:"
    echo "  -p, --port PORT     DNS port (default: $DEFAULT_PORT)"
    echo "  -c, --config FILE   Config file path (default: $DEFAULT_CONFIG)"
    echo "  -d, --daemon        Run as daemon"
    echo ""
    echo "Examples:"
    echo "  $0 deploy                    # Deploy with defaults"
    echo "  $0 deploy -p 53 -d          # Deploy on port 53 as daemon"
    echo "  $0 deploy -c custom.yaml    # Deploy with custom config"
}

deploy() {
    local port=$DEFAULT_PORT
    local config=$DEFAULT_CONFIG
    local daemon=false
    local run_args=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--port)
                port="$2"
                shift 2
                ;;
            -c|--config)
                config="$2"
                shift 2
                ;;
            -d|--daemon)
                daemon=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    print_info "Deploying PocketConcierge..."

    # Check if container already exists
    if docker ps -a --format '{{.Names}}' | grep -Eq "^${PROJECT_NAME}$"; then
        print_warning "Container $PROJECT_NAME already exists. Stopping and removing..."
        docker stop $PROJECT_NAME >/dev/null 2>&1 || true
        docker rm $PROJECT_NAME >/dev/null 2>&1 || true
    fi

    # Build image if it doesn't exist
    if ! docker images --format '{{.Repository}}:{{.Tag}}' | grep -Eq "^${PROJECT_NAME}:latest$"; then
        print_info "Building Docker image..."
        make docker-build
    fi

    # Prepare run arguments
    run_args="-p ${port}:8053/udp -p ${port}:8053/tcp --name $PROJECT_NAME"
    
    if [ "$daemon" = true ]; then
        run_args="$run_args -d --restart unless-stopped"
    else
        run_args="$run_args --rm -it"
    fi

    # Add config volume if custom config specified
    if [ "$config" != "$DEFAULT_CONFIG" ] && [ -f "$config" ]; then
        run_args="$run_args -v $(pwd)/$config:/app/config.yaml:ro"
    fi

    # Run container
    print_info "Starting container on port $port..."
    docker run $run_args $PROJECT_NAME:latest

    if [ "$daemon" = true ]; then
        print_success "PocketConcierge deployed as daemon on port $port"
        print_info "Use '$0 logs' to view logs"
        print_info "Use '$0 stop' to stop the service"
    fi
}

stop() {
    print_info "Stopping PocketConcierge..."
    docker stop $PROJECT_NAME >/dev/null 2>&1 || print_warning "Container not running"
    docker rm $PROJECT_NAME >/dev/null 2>&1 || print_warning "Container not found"
    print_success "PocketConcierge stopped"
}

restart() {
    print_info "Restarting PocketConcierge..."
    stop
    deploy -d
}

logs() {
    print_info "Showing PocketConcierge logs..."
    docker logs -f $PROJECT_NAME
}

status() {
    print_info "PocketConcierge status:"
    if docker ps --format '{{.Names}}' | grep -Eq "^${PROJECT_NAME}$"; then
        docker ps --filter "name=$PROJECT_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        print_success "PocketConcierge is running"
    else
        print_warning "PocketConcierge is not running"
    fi
}

clean() {
    print_info "Cleaning PocketConcierge resources..."
    docker stop $PROJECT_NAME >/dev/null 2>&1 || true
    docker rm $PROJECT_NAME >/dev/null 2>&1 || true
    docker rmi $PROJECT_NAME:latest >/dev/null 2>&1 || true
    print_success "PocketConcierge resources cleaned"
}

# Main script logic
case "${1:-help}" in
    deploy)
        shift
        deploy "$@"
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    logs)
        logs
        ;;
    status)
        status
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
