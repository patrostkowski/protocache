#!/usr/bin/env bash

set -euo pipefail

TLS_DIR="/etc/protocache/"
CERT_FILE="${TLS_DIR}/cert.pem"
KEY_FILE="${TLS_DIR}/key.pem"

mkdir -p "${TLS_DIR}"

echo "Generating TLS key..."
openssl genrsa -out "${KEY_FILE}" 2048

echo "Generating self-signed TLS certificate..."
openssl req -new -x509 -key "${KEY_FILE}" -out "${CERT_FILE}" -days 365 -subj "/CN=localhost"

echo "TLS certificate and key generated:"
echo "  Cert: ${CERT_FILE}"
echo "  Key:  ${KEY_FILE}"
