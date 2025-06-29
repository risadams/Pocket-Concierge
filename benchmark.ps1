Write-Host "🏁 DNS Performance Comparison" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "🏠 Testing PocketConcierge LOCAL resolution..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 127.0.0.1:8053 500 20 local

Write-Host ""
Write-Host "🌐 Testing PocketConcierge UPSTREAM forwarding..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 127.0.0.1:8053 500 20 upstream

Write-Host ""
Write-Host "🔀 Testing PocketConcierge MIXED queries..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 127.0.0.1:8053 500 20 mixed

Write-Host ""
Write-Host "📊 Testing Google DNS baseline..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 8.8.8.8:53 500 20 baseline

Write-Host ""
Write-Host "📊 Testing ControlD direct..." -ForegroundColor Yellow
go run cmd/benchmark/main.go 76.76.2.180:53 500 20 baseline

Write-Host ""
Write-Host "✅ Benchmark comparison complete!" -ForegroundColor Green
