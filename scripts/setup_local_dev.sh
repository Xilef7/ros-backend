#!/bin/bash
# setup_local_dev.sh: Reset and seed local development environment for Restaurant Ordering System
# Usage: ./scripts/setup_local_dev.sh
set -euo pipefail

# 1. Stop and remove existing containers
if docker compose ps -q | grep -q .; then
  echo "Stopping running containers..."
  docker compose down
fi

# 2. Remove old database volumes (clear persistent data)
echo "Removing old database volumes..."
docker volume rm $(docker volume ls -qf 'name=backend_postgres') 2>/dev/null || true

echo "Pruning dangling volumes..."
docker volume prune -f

# 3. Start containers with Docker Compose
echo "Starting containers..."
docker compose up -d --wait

# 4. Run database migrations (if migration tool exists)
if [ -f ./scripts/migrate.sh ]; then
  echo "Running DB migrations..."
  ./scripts/migrate.sh
else
  echo "No migration script found, skipping migrations."
fi

# 5. Insert seed data (if seed script exists)
if [ -f ./scripts/seed.sh ]; then
  echo "Seeding database..."
  ./scripts/seed.sh
else
  echo "No seed script found, skipping seeding."
fi

# 5. Generate TLS certificate (if tls script exists)
if [ -f ./scripts/generate_tls.sh ]; then
  echo "Generating certificate..."
  ./scripts/generate_tls.sh
else
  echo "No tls script found, skipping certificate."
fi

echo "Local development environment is ready."
