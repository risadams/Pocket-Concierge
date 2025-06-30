# PocketConcierge Docker Deployment Script for Windows PowerShell
# This script helps deploy PocketConcierge using Docker

param(
    [Parameter(Position=0)]
    [ValidateSet("deploy", "stop", "restart", "logs", "status", "clean", "help")]
    [string]$Command = "help",
    
    [Parameter()]
    [int]$Port = 8053,
    
    [Parameter()]
    [string]$Config = "config.yaml",
    
    [Parameter()]
    [switch]$Daemon
)

$PROJECT_NAME = "pocketconcierge"

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor $Colors.Red
}

function Show-Help {
    Write-Host "PocketConcierge Docker Deployment Script for Windows" -ForegroundColor $Colors.Blue
    Write-Host ""
    Write-Host "Usage: .\deploy.ps1 [COMMAND] [OPTIONS]" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor $Colors.White
    Write-Host "  deploy    Deploy PocketConcierge container" -ForegroundColor $Colors.White
    Write-Host "  stop      Stop PocketConcierge container" -ForegroundColor $Colors.White
    Write-Host "  restart   Restart PocketConcierge container" -ForegroundColor $Colors.White
    Write-Host "  logs      Show container logs" -ForegroundColor $Colors.White
    Write-Host "  status    Show container status" -ForegroundColor $Colors.White
    Write-Host "  clean     Remove container and image" -ForegroundColor $Colors.White
    Write-Host "  help      Show this help message" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor $Colors.White
    Write-Host "  -Port PORT      DNS port (default: 8053)" -ForegroundColor $Colors.White
    Write-Host "  -Config FILE    Config file path (default: config.yaml)" -ForegroundColor $Colors.White
    Write-Host "  -Daemon         Run as daemon" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor $Colors.White
    Write-Host "  .\deploy.ps1 deploy                    # Deploy with defaults" -ForegroundColor $Colors.White
    Write-Host "  .\deploy.ps1 deploy -Port 53 -Daemon  # Deploy on port 53 as daemon" -ForegroundColor $Colors.White
    Write-Host "  .\deploy.ps1 deploy -Config custom.yaml # Deploy with custom config" -ForegroundColor $Colors.White
}

function Deploy-Container {
    Write-Info "Deploying PocketConcierge..."

    # Check if container already exists
    $existingContainer = docker ps -a --format '{{.Names}}' | Where-Object { $_ -eq $PROJECT_NAME }
    if ($existingContainer) {
        Write-Warning "Container $PROJECT_NAME already exists. Stopping and removing..."
        docker stop $PROJECT_NAME 2>$null | Out-Null
        docker rm $PROJECT_NAME 2>$null | Out-Null
    }

    # Build image if it doesn't exist
    $existingImage = docker images --format '{{.Repository}}:{{.Tag}}' | Where-Object { $_ -eq "${PROJECT_NAME}:latest" }
    if (-not $existingImage) {
        Write-Info "Building Docker image..."
        make docker-build
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to build Docker image"
            exit 1
        }
    }

    # Prepare run arguments
    $runArgs = @(
        "-p", "${Port}:8053/udp",
        "-p", "${Port}:8053/tcp",
        "--name", $PROJECT_NAME
    )
    
    if ($Daemon) {
        $runArgs += @("-d", "--restart", "unless-stopped")
    } else {
        $runArgs += @("--rm", "-it")
    }

    # Add config volume if custom config specified
    if (($Config -ne "config.yaml") -and (Test-Path $Config)) {
        $configPath = (Resolve-Path $Config).Path
        $runArgs += @("-v", "${configPath}:/app/config.yaml:ro")
    }

    $runArgs += @("${PROJECT_NAME}:latest")

    # Run container
    Write-Info "Starting container on port $Port..."
    & docker run @runArgs

    if ($Daemon) {
        Write-Success "PocketConcierge deployed as daemon on port $Port"
        Write-Info "Use '.\deploy.ps1 logs' to view logs"
        Write-Info "Use '.\deploy.ps1 stop' to stop the service"
    }
}

function Stop-Container {
    Write-Info "Stopping PocketConcierge..."
    docker stop $PROJECT_NAME 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) {
        docker rm $PROJECT_NAME 2>$null | Out-Null
        Write-Success "PocketConcierge stopped"
    } else {
        Write-Warning "Container not running or not found"
    }
}

function Restart-Container {
    Write-Info "Restarting PocketConcierge..."
    Stop-Container
    $script:Daemon = $true
    Deploy-Container
}

function Show-Logs {
    Write-Info "Showing PocketConcierge logs..."
    docker logs -f $PROJECT_NAME
}

function Show-Status {
    Write-Info "PocketConcierge status:"
    $runningContainer = docker ps --format '{{.Names}}' | Where-Object { $_ -eq $PROJECT_NAME }
    if ($runningContainer) {
        docker ps --filter "name=$PROJECT_NAME" --format "table {{.Names}}`t{{.Status}}`t{{.Ports}}"
        Write-Host ""
        Write-Success "PocketConcierge is running"
    } else {
        Write-Warning "PocketConcierge is not running"
    }
}

function Clean-Resources {
    Write-Info "Cleaning PocketConcierge resources..."
    docker stop $PROJECT_NAME 2>$null | Out-Null
    docker rm $PROJECT_NAME 2>$null | Out-Null
    docker rmi "${PROJECT_NAME}:latest" 2>$null | Out-Null
    Write-Success "PocketConcierge resources cleaned"
}

# Main script logic
switch ($Command) {
    "deploy" {
        Deploy-Container
    }
    "stop" {
        Stop-Container
    }
    "restart" {
        Restart-Container
    }
    "logs" {
        Show-Logs
    }
    "status" {
        Show-Status
    }
    "clean" {
        Clean-Resources
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "Unknown command: $Command"
        Write-Host ""
        Show-Help
        exit 1
    }
}
