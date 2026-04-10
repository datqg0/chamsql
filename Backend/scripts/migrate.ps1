# Database migration and seed runner for Windows

# Load .env file
$env_file = ".\.env"
if (Test-Path $env_file) {
    Get-Content $env_file | Where-Object { $_ -notmatch "^#" -and $_ -ne "" } | ForEach-Object {
        $line = $_
        $key, $value = $line -split "=", 2
        [Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim())
    }
}

$DB_URI = $env:DATABASE_URI
if (-not $DB_URI) {
    $DB_URI = "postgres://postgres:123456@localhost:5432/sql_exam_db?sslmode=disable"
}

Write-Host "🔄 Running database migrations and seeding..." -ForegroundColor Cyan
Write-Host "Database URI: $DB_URI" -ForegroundColor Yellow

# Function to run SQL file using psql
function Run-SqlFile {
    param(
        [string]$FilePath
    )
    
    if (Test-Path $FilePath) {
        Write-Host "Running: $FilePath" -ForegroundColor Green
        $content = Get-Content $FilePath -Raw
        
        # Use psql to execute
        $content | psql -d $DB_URI
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ Error running $FilePath" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "⚠️  File not found: $FilePath" -ForegroundColor Yellow
    }
}

# Run all migration files in order
Write-Host "📦 Running schema migrations..." -ForegroundColor Cyan

$schema_files = @(
    "sql/schema/001_sql_exam_schema.sql",
    "sql/schema/002_add_refresh_tokens.sql",
    "sql/schema/003_add_test_cases.sql",
    "sql/schema/004_create_outbox_tables.sql",
    "sql/schema/005_create_permissions.sql",
    "sql/schema/006_create_classes.sql",
    "sql/schema/007_seed_test_data.sql"
)

foreach ($file in $schema_files) {
    Run-SqlFile $file
}

Write-Host "✅ Database setup complete!" -ForegroundColor Green
