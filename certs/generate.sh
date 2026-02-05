#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)

CA_DIR="$ROOT_DIR/ca"
SINK_DIR="$ROOT_DIR/sink"
NODE_DIR="$ROOT_DIR/node"

mkdir -p "$CA_DIR" "$SINK_DIR" "$NODE_DIR"

echo "===> Generating CA"
openssl genrsa -out "$CA_DIR/ca.key" 4096

openssl req -x509 -new -nodes \
  -key "$CA_DIR/ca.key" \
  -sha256 \
  -days 3650 \
  -out "$CA_DIR/ca.pem" \
  -subj "/C=US/O=Telemetry/CN=telemetry-ca"

echo "===> Generating Sink certificate"
openssl genrsa -out "$SINK_DIR/sink.key" 4096

openssl req -new \
  -key "$SINK_DIR/sink.key" \
  -out "$SINK_DIR/sink.csr" \
  -subj "/C=US/O=Telemetry/CN=telemetry-sink"

cat > "$SINK_DIR/sink.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = telemetry-sink
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

openssl x509 -req \
  -in "$SINK_DIR/sink.csr" \
  -CA "$CA_DIR/ca.pem" \
  -CAkey "$CA_DIR/ca.key" \
  -CAcreateserial \
  -out "$SINK_DIR/sink.pem" \
  -days 365 \
  -sha256 \
  -extfile "$SINK_DIR/sink.ext"

echo "===> Generating Node certificate"
openssl genrsa -out "$NODE_DIR/node.key" 4096

openssl req -new \
  -key "$NODE_DIR/node.key" \
  -out "$NODE_DIR/node.csr" \
  -subj "/C=US/O=Telemetry/CN=telemetry-node"

cat > "$NODE_DIR/node.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

openssl x509 -req \
  -in "$NODE_DIR/node.csr" \
  -CA "$CA_DIR/ca.pem" \
  -CAkey "$CA_DIR/ca.key" \
  -CAcreateserial \
  -out "$NODE_DIR/node.pem" \
  -days 365 \
  -sha256 \
  -extfile "$NODE_DIR/node.ext"

echo "✅ Certificates generated"
