# Blok2ttrpg Character Sheet - Build & Run (Dev Mode)
# Usage: ./start.ps1

# Kill any existing instance on port 8080
$existing = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
if ($existing) {
    Write-Host "Killing existing process on port 8080..." -ForegroundColor Yellow
    $existing | ForEach-Object { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue }
    Start-Sleep -Seconds 1
}

Write-Host "Building CSS..." -ForegroundColor Cyan
npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify
if ($LASTEXITCODE -ne 0) {
    Write-Host "CSS build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "Building Go binary..." -ForegroundColor Cyan
Remove-Item -Path bin -Recurse -Force -ErrorAction SilentlyContinue
go build -o bin/charsheet.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Go build failed!" -ForegroundColor Red
    exit 1
}

$port = if ($env:PORT) { $env:PORT } else { "8080" }

# DEV=1 makes the server load templates from disk (not embed)
# so you can edit templates and just refresh the browser
$env:DEV = "1"

Write-Host ""
Write-Host "DEV MODE - templates loaded from disk (edit & refresh)" -ForegroundColor Magenta
Write-Host "Starting server at http://localhost:$port" -ForegroundColor Green
Write-Host "Press Ctrl+C to stop" -ForegroundColor Gray
Write-Host ""

./bin/charsheet.exe
