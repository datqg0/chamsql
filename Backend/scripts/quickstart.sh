#!/bin/bash
# Quick start script for testing ChamsQL Backend

set -e

echo "🚀 ChamsQL Backend - Quick Start"
echo "================================"
echo ""

# Step 1: Start Docker
echo "📦 Step 1: Starting PostgreSQL Docker container..."
docker-compose up -d
echo "✅ Database is running on localhost:5432"
echo ""

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
until docker exec sqlexam.postgres pg_isready -U postgres > /dev/null 2>&1; do
    echo "  Database not ready yet, waiting..."
    sleep 2
done
echo "✅ Database is ready!"
echo ""

# Step 2: Run migrations and seed data
echo "🔄 Step 2: Running migrations and seeding test data..."
if command -v psql &> /dev/null; then
    bash scripts/migrate.sh
    echo "✅ Database initialized with test data!"
else
    echo "⚠️  psql not found. You need to run migrations manually:"
    echo "   Windows: .\scripts\migrate.ps1"
    echo "   Linux/Mac: bash scripts/migrate.sh"
    echo ""
fi
echo ""

# Step 3: Build application
echo "🔨 Step 3: Building application..."
go build -o app.exe ./cmd/app/main.go
echo "✅ Application built successfully!"
echo ""

# Step 4: Start application
echo "🎯 Step 4: Starting backend server..."
echo "   API: http://localhost:8080"
echo "   Documentation: See TESTING_GUIDE.md"
echo ""
go run ./cmd/app/main.go
