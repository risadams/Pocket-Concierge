# Extract port from config.yaml
$ConfigPath = "config.yaml"
if (-not (Test-Path $ConfigPath)) {
    $ConfigPath = "configs/example.yaml"
}

$Port = 8053  # Default fallback
if (Test-Path $ConfigPath) {
    try {
        $ConfigContent = Get-Content $ConfigPath -Raw
        if ($ConfigContent -match "server:\s*\n\s*port:\s*(\d+)") {
            $Port = [int]$Matches[1]
        }
    }
    catch {
        Write-Warning "Could not parse config file, using default port 8053"
    }
}

Write-Host "🏁 DNS Performance Comparison" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "📋 Using port: $Port" -ForegroundColor White

Write-Host ""
Write-Host "🏠 Testing PocketConcierge LOCAL resolution..." -ForegroundColor Yellow
go run cmd/benchmark/main.go "127.0.0.1:$Port" 500 20 local

Write-Host ""
Write-Host "🌐 Testing PocketConcierge UPSTREAM forwarding..." -ForegroundColor Yellow
go run cmd/benchmark/main.go "127.0.0.1:$Port" 500 20 upstream

Write-Host ""
Write-Host "🔀 Testing PocketConcierge MIXED queries..." -ForegroundColor Yellow
go run cmd/benchmark/main.go "127.0.0.1:$Port" 500 20 mixed

Write-Host ""
Write-Host "📊 Testing Google DNS baseline..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 8.8.8.8:53 500 20 baseline

Write-Host ""
Write-Host "📊 Testing ControlD direct..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 76.76.2.180:53 500 20 baseline

Write-Host ""
Write-Host "✅ Benchmark comparison complete!" -ForegroundColor Green
