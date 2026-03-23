# Architecture Documentation

## Overview

The Avalanche Remote Signer is designed with a modular, pluggable architecture that separates validator signing operations from key management.

## Components

### 1. gRPC Server (`pkg/server`)

Exposes a gRPC API that AvalancheGo connects to for signing operations.

**API Methods:**
- `Sign(SignRequest) → SignResponse` - Signs messages with BLS
- `GetPublicKey(PublicKeyRequest) → PublicKeyResponse` - Returns validator public key
- `HealthCheck(HealthCheckRequest) → HealthCheckResponse` - Service health status

**Features:**
- TLS support for secure communication
- Request validation
- Structured logging
- Graceful shutdown

### 2. BLS Signer (`pkg/signer`)

Core signing logic that implements Avalanche's BLS12-381 signature scheme.

**Responsibilities:**
- Message signing with BLS
- Public key retrieval
- Signature verification (for testing)
- Backend abstraction

### 3. Backend Interface (`pkg/signer/backend.go`)

Defines the contract that all key management backends must implement:

```go
type SignerBackend interface {
    Sign(ctx context.Context, message []byte) ([]byte, error)
    GetPublicKey(ctx context.Context) ([]byte, error)
    Close() error
}
```

This allows plugging in different storage solutions without changing core logic.

### 4. Backend Implementations

#### File Backend (`pkg/backend/file`)

Simple file-based key storage for development and testing.

**⚠️ NOT FOR PRODUCTION**

**Features:**
- JSON key file format
- File permission checks (600)
- In-memory key caching
- Key generation utility

#### Vault Backend (Planned)

HashiCorp Vault integration for production use.

**Features:**
- Secure secret storage
- Access control policies
- Audit logging
- Key rotation support

#### AWS KMS Backend (Planned)

AWS Key Management Service integration.

**Features:**
- Cloud-native key management
- IAM-based access control
- Multi-region support
- Hardware security modules (HSM)

#### Google Cloud KMS Backend (Planned)

Google Cloud KMS integration.

#### Azure Key Vault Backend (Planned)

Microsoft Azure Key Vault integration.

### 5. Configuration (`pkg/config`)

Centralized configuration management using Viper.

**Sources:**
1. Configuration file (YAML)
2. Environment variables (prefix: `AVAX_SIGNER_`)
3. Defaults

**Example:**
```yaml
server:
  address: "0.0.0.0:9090"
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"

backend:
  type: "vault"
  config:
    address: "https://vault.example.com:8200"
    token: "${VAULT_TOKEN}"

logging:
  level: "info"
  format: "json"
```

### 6. CLI (`cmd/signer`)

Command-line interface built with Cobra.

**Commands:**
- `avax-remote-signer` - Start the signer server
- `avax-remote-signer genkey` - Generate BLS key pair

## Data Flow

### Signing Flow

```
┌──────────────┐
│ AvalancheGo  │
└──────┬───────┘
       │ 1. gRPC SignRequest
       │    (message bytes)
       ▼
┌──────────────┐
│ gRPC Server  │
└──────┬───────┘
       │ 2. SignMessage()
       ▼
┌──────────────┐
│  BLS Signer  │
└──────┬───────┘
       │ 3. Sign()
       ▼
┌──────────────┐
│   Backend    │
│  (Vault/KMS) │
└──────┬───────┘
       │ 4. BLS signature
       │
       ▼
    [return path]
```

**Steps:**
1. AvalancheGo sends consensus message to signer via gRPC
2. gRPC server validates request and calls BLS signer
3. BLS signer delegates to configured backend
4. Backend retrieves key and performs BLS signature
5. Signature bubbles back up to AvalancheGo

### Key Retrieval Flow

```
┌──────────────┐
│ AvalancheGo  │
└──────┬───────┘
       │ GetPublicKeyRequest
       ▼
┌──────────────┐
│ gRPC Server  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  BLS Signer  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Backend    │
│  (Cache)     │
└──────┬───────┘
       │ BLS public key
       ▼
    [return]
```

Public keys are typically cached for performance.

## Security Model

### Threat Model

**Assets:**
- BLS private key (validator signing key)

**Threats:**
- Key theft from validator node
- Unauthorized signing operations
- Man-in-the-middle attacks
- Key exposure in logs/memory dumps

**Mitigations:**
1. **Key Isolation**: Keys stored in separate system (Vault/KMS)
2. **TLS**: Encrypted communication between validator and signer
3. **Network Isolation**: Firewall rules restrict access
4. **Access Control**: KMS/Vault policies limit key access
5. **Audit Logging**: All operations logged
6. **Memory Safety**: Minimize key lifetime in memory

### Trust Boundaries

```
┌─────────────────────────────────────────────────┐
│ Untrusted Network                                │
└─────────────────────────────────────────────────┘
                    │ TLS
┌─────────────────────────────────────────────────┐
│ Validator Host (Semi-trusted)                   │
│  ┌──────────────┐                               │
│  │ AvalancheGo  │                               │
│  └──────────────┘                               │
└─────────────────────────────────────────────────┘
                    │ TLS/gRPC
┌─────────────────────────────────────────────────┐
│ Signer Host (Trusted)                           │
│  ┌──────────────┐                               │
│  │   Signer     │                               │
│  └──────────────┘                               │
└─────────────────────────────────────────────────┘
                    │ Authenticated API
┌─────────────────────────────────────────────────┐
│ KMS (Highly Trusted)                            │
│  ┌──────────────┐                               │
│  │  Private Key │                               │
│  └──────────────┘                               │
└─────────────────────────────────────────────────┘
```

## Performance Considerations

### Latency

Signing operations must complete quickly to avoid consensus delays.

**Target:** < 100ms per signature

**Optimization strategies:**
- Connection pooling to backend
- Public key caching
- Async logging
- Minimal allocations

### Throughput

Validators may need to sign many messages per second during consensus.

**Target:** > 100 signatures/second

**Scaling:**
- Multiple signer instances behind load balancer
- Backend-specific optimizations (local caching, etc.)

## Error Handling

### Error Categories

1. **Transient Errors**: Retry with backoff
   - Network timeouts
   - Backend unavailable

2. **Permanent Errors**: Return error to validator
   - Invalid message format
   - Key not found

3. **Critical Errors**: Alert and investigate
   - Backend corruption
   - Security policy violation

### Failure Modes

**Backend Unavailable:**
- Retry with exponential backoff
- Alert operators
- Validator fails to sign → misses consensus rounds

**TLS Certificate Expired:**
- Signer rejects connections
- Validator cannot sign
- Monitoring should catch before expiry

**Key Corruption:**
- Backend returns invalid signature
- Validator rejects signature
- Manual intervention required

## Extensibility

### Adding New Backends

1. Implement `SignerBackend` interface
2. Add configuration parsing
3. Register in `createBackend()` function
4. Add tests
5. Document in `docs/BACKENDS.md`

### Adding New Features

Examples of future enhancements:

- **Multi-key support**: Sign with different keys for different validators
- **Rate limiting**: Prevent DoS on signing operations
- **Metrics export**: Prometheus metrics endpoint
- **Key rotation**: Automated key rotation workflows
- **HSM support**: Direct hardware security module integration

## Testing Strategy

### Unit Tests
- Test each component in isolation
- Mock dependencies
- Fast execution (< 1s)

### Integration Tests
- Test component interactions
- Use test containers for backends
- Medium execution time (< 30s)

### End-to-End Tests
- Full system test with AvalancheGo
- Verify actual signing operations
- Slow execution (minutes)

### Security Tests
- Penetration testing
- Fuzzing inputs
- Static analysis
- Dependency scanning

## Deployment Patterns

### Pattern 1: Sidecar

Signer runs on same host as validator.

**Pros:** Low latency
**Cons:** Less security isolation

### Pattern 2: Remote Signer

Signer on dedicated host.

**Pros:** Better security
**Cons:** Network dependency

### Pattern 3: HA Setup

Multiple signers behind load balancer.

**Pros:** High availability
**Cons:** Complexity

See [PRODUCTION.md](PRODUCTION.md) for detailed deployment guidance.

## Monitoring

### Key Metrics

- `avax_signer_sign_requests_total` - Total sign requests
- `avax_signer_sign_duration_seconds` - Sign latency histogram
- `avax_signer_sign_errors_total` - Sign errors by type
- `avax_signer_backend_status` - Backend health (0/1)

### Health Checks

- gRPC health check endpoint
- Backend connectivity check
- Key accessibility check

### Alerting

Critical alerts:
- Signing failures > threshold
- Backend unavailable
- High latency (> 500ms)
- Certificate expiring soon
- Process crash/restart

## References

- [Avalanche Documentation](https://docs.avax.network/)
- [BLS Signatures](https://en.wikipedia.org/wiki/BLS_digital_signature)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [12-Factor App](https://12factor.net/)
