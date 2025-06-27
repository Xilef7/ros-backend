#!/bin/sh
# Generate a self-signed TLS certificate
# Usage: ./generate_tls.sh

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="$SCRIPT_DIR/../certs"
CERT_FILE="server_cert.pem"
KEY_FILE="server_key.pem"

mkdir "$CERT_DIR"

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
-keyout "$CERT_DIR/$KEY_FILE" -out "$CERT_DIR/$CERT_FILE" \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost"