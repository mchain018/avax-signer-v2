# Quick Start Guide

This guide will help you get the Avalanche Remote Signer running quickly.

## Prerequisites

- Go 1.23 or later
- AvalancheGo node (for integration)
- Basic understanding of Avalanche validators

## Installation

### Build from source

```bash
git clone <your-repo-url>
cd avax-remote-signer
go build -o avax-remote-signer ./cmd/signer
```

### Verify installation

```bash
./avax-remote-signer --help
```

## Development Setup (File Backend)

⚠️ **WARNING**: File backend is for DEVELOPMENT/TESTING only. Never use in production!

### Step 1: Generate a BLS key pair

```bash
./avax-remote-signer genkey -o ./keys/validator-key.json
```

This creates a new BLS key pair and saves it to `./keys/validator-key.json`.

### Step 2: Create configuration

Copy the example config:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
server:
  address: "0.0.0.0:9090"
  tls:
    enabled: false  # Disable for local testing only

backend:
  type: "file"
  config:
    key_path: "./keys/validator-key.json"

logging:
  level: "debug"
  format: "console"
```

### Step 3: Start the signer

```bash
./avax-remote-signer --config config.yaml
```

You should see output like:

```
INFO    Starting Avalanche Remote Signer    {"backend": "file", "address": "0.0.0.0:9090"}
WARN    Using file-based backend - NOT RECOMMENDED FOR PRODUCTION
INFO    Validator public key loaded    {"public_key": "a1b2c3..."}
INFO    Starting gRPC server    {"address": "0.0.0.0:9090"}
```

### Step 4: Configure AvalancheGo

Add to your AvalancheGo node configuration:

```json
{
  "remote-signer-enabled": true,
  "remote-signer-url": "grpc://localhost:9090"
}
```

Restart your AvalancheGo node to connect to the remote signer.

## Production Setup

For production, you should use a proper key management system:

### Option 1: HashiCorp Vault

1. Store your BLS key in Vault
2. Configure the signer:

```yaml
backend:
  type: "vault"
  config:
    address: "https://vault.example.com:8200"
    token: "${VAULT_TOKEN}"
    key_path: "secret/data/avalanche/validator"
```

### Option 2: AWS KMS

1. Create a KMS key for BLS signing
2. Configure the signer:

```yaml
backend:
  type: "awskms"
  config:
    region: "us-east-1"
    key_id: "arn:aws:kms:..."
```

### Always use TLS in production

Generate certificates:

```bash
# Generate self-signed cert for testing
openssl req -x509 -newkey rsa:4096 -keyout server-key.pem -out server-cert.pem -days 365 -nodes
```

Update config:

```yaml
server:
  tls:
    enabled: true
    cert_file: "/path/to/server-cert.pem"
    key_file: "/path/to/server-key.pem"
```

## Testing

Run the test suite:

```bash
go test ./...
```

## Troubleshooting

### "failed to load key file"

- Check that the key path in config.yaml is correct
- Verify the key file exists and is readable
- Check file permissions (should be 600)

### "failed to listen on address"

- Check if port 9090 is already in use: `lsof -i :9090`
- Try a different port in your config
- Ensure you have permission to bind to the address

### "TLS handshake error"

- Verify cert and key files are valid
- Check that AvalancheGo is configured to accept the certificate
- For testing, you can disable TLS (not recommended)

## Next Steps

- Set up monitoring and alerting
- Configure log aggregation
- Implement key rotation procedures
- Review security hardening checklist
- Test failover scenarios

## Security Checklist

- [ ] Use TLS in production
- [ ] Run on isolated network segment
- [ ] Configure firewall rules (only AvalancheGo should access signer)
- [ ] Use proper key management system (Vault, KMS, HSM)
- [ ] Enable audit logging
- [ ] Set up monitoring and alerting
- [ ] Document incident response procedures
- [ ] Regular security audits
- [ ] Keep software updated

## Support

For issues and questions:
- GitHub Issues: <your-repo-url>/issues
- Documentation: <your-docs-url>
