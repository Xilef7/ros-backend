#!/bin/bash
# scripts/seed.sh: Insert seed data into the local development database
set -euo pipefail

# Example: use a Go CLI command to seed the database
# You can replace this with your preferred seeding method

if [ -f ./cmd/cli/main.go ]; then
  echo "Seeding database using Go CLI..."
  go run ./cmd/cli/main.go seed
else
  echo "No seed entrypoint found. Please implement seed logic."
  exit 1
fi
