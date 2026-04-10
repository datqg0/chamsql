# Quick start script for testing ChamsQL Backend (Windows)

Write-Host "🚀 ChamsQL Backend - Quick Start (Windows)" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Start Docker
Write-Host "📦 Step 1: Starting PostgreSQL Docker container..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Database is running on localhost:5432" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to start Docker containers" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Wait for database to be ready
Write-Host "⏳ Waiting for database to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
while ($attempt -lt $maxAttempts) {
    try {
        $result = docker exec sqlexam.postgres pg_isready -U postgres 2>&1
        if ($LASTEXITCODE -eq 0) {
            break
        }
    } catch {
        # Continue waiting
    }
    $attempt++
    if ($attempt -lt $maxAttempts) {
        Write-Host "  Database not ready yet (attempt $attempt/$maxAttempts), waiting..." -ForegroundColor Gray
        Start-Sleep -Seconds 2
    }
}

if ($attempt -ge $maxAttempts) {
    Write-Host "❌ Database did not become ready in time" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Database is ready!" -ForegroundColor Green
Write-Host ""

# Step 2: Run migrations and seed data
Write-Host "🔄 Step 2: Running migrations and seeding test data..." -ForegroundColor Yellow
if (Test-Path "scripts/migrate.ps1") {
    & .\scripts\migrate.ps1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Database initialized with test data!" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Migration had issues, but continuing..." -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  Migration script not found. Please run manually:" -ForegroundColor Yellow
    Write-Host "   .\scripts\migrate.ps1" -ForegroundColor Yellow
}
Write-Host ""

# Step 3: Build application
Write-Host "🔨 Step 3: Building application..." -ForegroundColor Yellow
go build -o app.exe ./cmd/app/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Application built successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 4: Start application
Write-Host "🎯 Step 4: Starting backend server..." -ForegroundColor Yellow
Write-Host "   API: http://localhost:8080" -ForegroundColor Cyan
Write-Host "   Documentation: See TESTING_GUIDE.md" -ForegroundColor Cyan
Write-Host ""

go run ./cmd/app/main.go
