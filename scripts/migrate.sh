#!/bin/bash
# scripts/migrate.sh: Run database migrations for local development
set -euo pipefail

# Example: use GORM's AutoMigrate via a Go script
# You can replace this with your migration tool (e.g., golang-migrate, goose, etc.)

if [ -f ./cmd/cli/main.go ]; then
  echo "Running Go-based migrations (AutoMigrate)..."
  go run ./cmd/cli/main.go migrate
else
  echo "No migration entrypoint found. Please implement migration logic."
  exit 1
fi
