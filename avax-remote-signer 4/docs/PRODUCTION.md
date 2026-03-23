# Production Deployment Guide

This guide covers deploying the Avalanche Remote Signer in production environments.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Secure Network Zone                       │
│                                                               │
│  ┌──────────────┐         TLS/gRPC        ┌──────────────┐ │
│  │              │ ◀─────────────────────▶ │              │ │
│  │ AvalancheGo  │                          │   Remote     │ │
│  │   Validator  │                          │   Signer     │ │
│  │              │                          │              │ │
│  └──────────────┘                          └──────┬───────┘ │
│                                                   │          │
└───────────────────────────────────────────────────┼──────────┘
                                                    │
                                         Authenticated
                                              API
                                                    │
┌───────────────────────────────────────────────────▼──────────┐
│                  Key Management Service                       │
│                                                               │
│  • HashiCorp Vault                                           │
│  • AWS KMS                                                   │
│  • Google Cloud KMS                                          │
│  • Azure Key Vault                                           │
│  • Hardware Security Module (HSM)                            │
└───────────────────────────────────────────────────────────────┘
```

## Deployment Options

### 1. Same Host as Validator (Not Recommended)

If the signer runs on the same host as the validator, you lose the security benefit of key isolation.

**Use case**: Small testing environments only

### 2. Separate Host (Recommended)

Run the signer on a dedicated, hardened host with restricted access.

**Architecture**:
- Validator host: AvalancheGo only
- Signer host: Remote signer + connection to KMS
- Network: Dedicated VLAN or VPN tunnel

### 3. High Availability Setup

For critical validators, deploy multiple signer instances:

```
                    ┌──────────────┐
                    │ Load Balancer│
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
    ┌───▼────┐        ┌────▼───┐        ┌────▼───┐
    │Signer 1│        │Signer 2│        │Signer 3│
    └───┬────┘        └────┬───┘        └────┬───┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                    ┌──────▼───────┐
                    │     KMS      │
                    └──────────────┘
```

## Key Management Backend Setup

### HashiCorp Vault

#### 1. Install Vault

```bash
# Using package manager
brew install vault  # macOS
apt-get install vault  # Ubuntu

# Or download binary
wget https://releases.hashicorp.com/vault/...
```

#### 2. Initialize Vault

```bash
vault server -dev  # Development mode

# Production: Use proper storage backend (Consul, etc.)
```

#### 3. Store validator key

```bash
# Enable KV secrets engine
vault secrets enable -path=secret kv-v2

# Store the key
vault kv put secret/avalanche/validator-key \
  private_key="$(cat validator-key.json | jq -r .private_key)" \
  public_key="$(cat validator-key.json | jq -r .public_key)"
```

#### 4. Create policy

```hcl
# validator-signer-policy.hcl
path "secret/data/avalanche/validator-key" {
  capabilities = ["read"]
}
```

```bash
vault policy write validator-signer validator-signer-policy.hcl
```

#### 5. Configure signer

```yaml
backend:
  type: "vault"
  config:
    address: "https://vault.example.com:8200"
    token: "${VAULT_TOKEN}"  # Use app role in production
    key_path: "secret/data/avalanche/validator-key"
```

### AWS KMS

#### 1. Create KMS key

```bash
aws kms create-key \
  --description "Avalanche Validator BLS Key" \
  --key-usage SIGN_VERIFY \
  --customer-master-key-spec ECC_SECG_P256K1
```

#### 2. Import existing key (if applicable)

```bash
# Get import parameters
aws kms get-parameters-for-import \
  --key-id <key-id> \
  --wrapping-algorithm RSAES_OAEP_SHA_256 \
  --wrapping-key-spec RSA_2048

# Wrap your key (use provided wrapping key)
# ... wrapping process ...

# Import wrapped key
aws kms import-key-material \
  --key-id <key-id> \
  --encrypted-key-material file://WrappedKey.bin \
  --import-token file://ImportToken.bin
```

#### 3. Configure IAM permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:Sign",
        "kms:GetPublicKey"
      ],
      "Resource": "arn:aws:kms:us-east-1:123456789012:key/..."
    }
  ]
}
```

#### 4. Configure signer

```yaml
backend:
  type: "awskms"
  config:
    region: "us-east-1"
    key_id: "arn:aws:kms:us-east-1:123456789012:key/..."
    # Uses AWS credentials from environment/IAM role
```

## Network Security

### Firewall Rules

```bash
# Allow only validator host to access signer
# Signer host (incoming)
ufw allow from <validator-ip> to any port 9090 proto tcp
ufw deny 9090/tcp

# Validator host (outgoing)
ufw allow out to <signer-ip> port 9090 proto tcp
```

### TLS Configuration

#### Generate production certificates

Option 1: Use Let's Encrypt (if publicly accessible)

```bash
certbot certonly --standalone \
  -d signer.example.com \
  --agree-tos \
  --email admin@example.com
```

Option 2: Use internal CA

```bash
# Generate CA
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -days 3650 -key ca-key.pem -out ca-cert.pem

# Generate server cert
openssl genrsa -out server-key.pem 4096
openssl req -new -key server-key.pem -out server.csr
openssl x509 -req -days 365 -in server.csr \
  -CA ca-cert.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-cert.pem
```

## Monitoring

### Metrics

Expose Prometheus metrics (add to server):

```go
// Add metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

Key metrics to monitor:
- Sign operation latency
- Sign operation success/failure rate
- Backend connection status
- gRPC connection count
- CPU and memory usage

### Logging

Configure structured logging:

```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/avax-signer/signer.log"
```

Use log aggregation (ELK, Splunk, Datadog):

```bash
# Forward logs to aggregator
filebeat -e -c filebeat.yml
```

### Alerting

Set up alerts for:
- Signing failures
- Backend connection loss
- High latency
- Process crashes
- Certificate expiration

Example Prometheus alert:

```yaml
groups:
- name: avalanche-signer
  rules:
  - alert: SignerHighLatency
    expr: signer_sign_duration_seconds > 0.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Signer latency is high"
```

## Backup and Recovery

### Key Backup

**Critical**: Always have secure backups of your validator keys!

```bash
# Backup from Vault
vault kv get -format=json secret/avalanche/validator-key > backup.json

# Encrypt backup
gpg --encrypt --recipient admin@example.com backup.json

# Store in secure location (offline, multiple locations)
```

### Disaster Recovery

1. **RTO (Recovery Time Objective)**: < 5 minutes
2. **RPO (Recovery Point Objective)**: 0 (keys don't change)

Recovery procedure:
1. Deploy new signer instance
2. Restore key from backup to KMS
3. Update configuration
4. Point validator to new signer
5. Verify signing operations

## Performance Tuning

### Signer Settings

```yaml
server:
  max_concurrent_requests: 100
  connection_timeout: 30s
  request_timeout: 10s
```

### System Settings

```bash
# Increase file descriptors
ulimit -n 65536

# Optimize network
sysctl -w net.core.rmem_max=134217728
sysctl -w net.core.wmem_max=134217728
```

## Security Hardening

### Host Hardening

```bash
# Update system
apt update && apt upgrade -y

# Configure firewall (see above)

# Disable unnecessary services
systemctl disable <unused-service>

# Configure fail2ban
apt install fail2ban

# Enable automatic security updates
apt install unattended-upgrades
```

### Application Security

- Run as non-root user
- Use systemd service with restrictions
- Enable SELinux/AppArmor
- Regular security audits
- Keep dependencies updated

### Example systemd service

```ini
[Unit]
Description=Avalanche Remote Signer
After=network.target

[Service]
Type=simple
User=avax-signer
Group=avax-signer
ExecStart=/usr/local/bin/avax-remote-signer --config /etc/avax-signer/config.yaml
Restart=always
RestartSec=10

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/avax-signer

[Install]
WantedBy=multi-user.target
```

## Maintenance

### Key Rotation

1. Generate new key pair
2. Store in KMS
3. Update signer configuration
4. Restart signer
5. Update validator with new public key
6. Wait for network to recognize change
7. Archive old key (don't delete immediately)

### Updates

```bash
# Test in staging first
./test-update.sh

# Backup current version
cp /usr/local/bin/avax-remote-signer /usr/local/bin/avax-remote-signer.backup

# Deploy new version
cp avax-remote-signer /usr/local/bin/

# Restart service
systemctl restart avax-remote-signer

# Monitor logs
journalctl -u avax-remote-signer -f
```

## Compliance

- Keep audit logs for required retention period
- Regular security assessments
- Document access control policies
- Maintain incident response plan
- Conduct tabletop exercises

## Support

For production support:
- GitHub Issues: <repo-url>/issues
- Email: support@example.com
- Discord: <discord-link>
