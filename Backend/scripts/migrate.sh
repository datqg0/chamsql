#!/bin/bash
# Database migration and seed runner

set -e

# Load environment from .env
export $(grep -v '^#' .env | xargs)

# Extract connection details from DATABASE_URI
# Expected format: postgres://user:password@host:port/dbname?sslmode=disable
DB_URI="${DATABASE_URI:-postgres://postgres:123456@localhost:5432/sql_exam_db?sslmode=disable}"

echo "🔄 Running database migrations and seeding..."
echo "Database URI: $DB_URI"

# Function to run SQL file
run_sql_file() {
    local file=$1
    echo "Running: $file"
    psql "$DB_URI" -f "$file"
}

# Run all migration files in order
echo "📦 Running schema migrations..."
run_sql_file "sql/schema/001_sql_exam_schema.sql"
run_sql_file "sql/schema/002_add_refresh_tokens.sql"
run_sql_file "sql/schema/003_add_test_cases.sql"
run_sql_file "sql/schema/004_create_outbox_tables.sql"
run_sql_file "sql/schema/005_create_permissions.sql"
run_sql_file "sql/schema/006_create_classes.sql"

# Seed test data
echo "🌱 Seeding test data..."
run_sql_file "sql/schema/007_seed_test_data.sql"

echo "✅ Database setup complete!"
