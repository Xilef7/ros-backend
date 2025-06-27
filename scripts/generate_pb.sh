#!/bin/sh
# Regenerate Go and gRPC code from restaurant.proto
# Usage: ./generate_pb.sh

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROTO_DIR="$SCRIPT_DIR/../api/proto"
PROTO_FILE="restaurant.proto"

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$PROTO_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$PROTO_DIR" \
  --go-grpc_opt=paths=source_relative \
  $PROTO_FILE

echo "Proto files regenerated."
