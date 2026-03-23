# Avalanche Remote Signer

An open-source remote signing solution for Avalanche validators with pluggable key management backends.

## Overview

This sidecar service sits between AvalancheGo and your key management system, providing BLS signature operations for validator consensus while keeping private keys isolated from the validator node.

## Features

- **Open Source**: Fully transparent, auditable code under BSD-3 license
- **Pluggable Backends**: Support for multiple key management systems:
  - File-based (development/testing)
  - HashiCorp Vault (roadmap)
  - Cloud KMS providers (roadmap)
  - Hardware Security Modules (roadmap)
- **BLS12-381 Signatures**: Native support for Avalanche's consensus signature scheme
- **gRPC API**: Compatible with AvalancheGo's remote signer interface

## Architecture

```
┌──────────────┐         gRPC          ┌──────────────┐
│              │ ───────────────────▶  │   Remote     │
│ AvalancheGo  │                        │   Signer     │
│              │ ◀───────────────────   │   Sidecar    │
└──────────────┘     BLS Signatures    └──────┬───────┘
                                              │
                                              │ Plugin Interface
                                              │
                                    ┌─────────▼──────────┐
                                    │   Key Management   │
                                    │      Backend       │
                                    │ (Vault, KMS, etc.) │
                                    └────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23 or later
- AvalancheGo node
- Key management backend (file, Vault, etc.)

### Installation

```bash
go build -o avax-remote-signer ./cmd/signer
```

### Configuration

Create a `config.yaml`:

```yaml
server:
  address: "0.0.0.0:9090"
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"

backend:
  type: "file"  # or "vault", "awskms", etc.
  config:
    key_path: "/path/to/validator-key.json"

logging:
  level: "info"
```

### Running

```bash
./avax-remote-signer --config config.yaml
```

## Supported Backends

### File Backend (Development)

For testing and development. **NOT recommended for production.**

```yaml
backend:
  type: "file"
  config:
    key_path: "/secure/path/to/key.json"
```

### HashiCorp Vault (Roadmap)

Enterprise-grade secret management.

```yaml
backend:
  type: "vault"
  config:
    address: "https://vault.example.com:8200"
    token: "${VAULT_TOKEN}"
    key_path: "secret/data/validator/key"
```

### AWS KMS (Roadmap)

Cloud-native key management.

### Google Cloud KMS (Roadmap)

Cloud-native key management.

## Configuring AvalancheGo

Configure your AvalancheGo node to use the remote signer:

```json
{
  "remote-signer-enabled": true,
  "remote-signer-url": "grpc://localhost:9090",
  "remote-signer-tls-cert": "/path/to/signer-cert.pem"
}
```

## Security Considerations

- **TLS Required**: Always use TLS in production
- **Network Isolation**: Run signer on isolated network segment
- **Access Control**: Implement strict firewall rules
- **Key Rotation**: Regular key rotation practices
- **Audit Logging**: Enable comprehensive audit logs

## Development

### Building from Source

```bash
git clone https://github.com/yourusername/avax-remote-signer
cd avax-remote-signer
go build ./cmd/signer
```

### Running Tests

```bash
go test ./...
```

### Adding a New Backend

Implement the `SignerBackend` interface in `pkg/backend/interface.go`:

```go
type SignerBackend interface {
    Sign(ctx context.Context, publicKey []byte, message []byte) ([]byte, error)
    GetPublicKey(ctx context.Context, keyID string) ([]byte, error)
}
```

## Roadmap

- [x] Core signer architecture
- [x] File-based backend
- [ ] HashiCorp Vault backend
- [ ] AWS KMS backend
- [ ] Google Cloud KMS backend
- [ ] HSM support
- [ ] Key rotation automation
- [ ] Metrics and monitoring
- [ ] Rate limiting
- [ ] Multi-key support for multiple validators

## License

BSD-3-Clause License - see [LICENSE](LICENSE) for details

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## Security

Report security vulnerabilities to security@example.com

## Acknowledgments

Inspired by Web3Signer and the Cubesigner sidecar architecture.
