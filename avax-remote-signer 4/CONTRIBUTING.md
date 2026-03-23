# Contributing to Avalanche Remote Signer

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Make (optional, but recommended)

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/avax-remote-signer.git
cd avax-remote-signer
```

### Branch Strategy

- `main` - stable release branch
- `develop` - integration branch for features
- `feature/*` - feature branches
- `fix/*` - bug fix branches

Create a feature branch:

```bash
git checkout -b feature/my-new-feature
```

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Lint

```bash
make lint
```

### Format Code

```bash
make fmt
```

## Adding a New Backend

To add support for a new key management backend:

### 1. Create backend package

```bash
mkdir -p pkg/backend/mybackend
```

### 2. Implement the interface

Create `pkg/backend/mybackend/mybackend.go`:

```go
package mybackend

import (
    "context"
    "github.com/avax-remote-signer/signer/pkg/signer"
)

type MyBackend struct {
    // your fields
}

func New(config map[string]interface{}) (signer.SignerBackend, error) {
    // initialization
}

func (b *MyBackend) Sign(ctx context.Context, message []byte) ([]byte, error) {
    // implement signing
}

func (b *MyBackend) GetPublicKey(ctx context.Context) ([]byte, error) {
    // implement public key retrieval
}

func (b *MyBackend) Close() error {
    // cleanup
}
```

### 3. Add tests

Create `pkg/backend/mybackend/mybackend_test.go`:

```go
package mybackend

import (
    "context"
    "testing"
)

func TestMyBackend(t *testing.T) {
    // test implementation
}
```

### 4. Register backend

Update `cmd/signer/main.go` in the `createBackend` function:

```go
case "mybackend":
    return mybackend.New(cfg.Config)
```

### 5. Document

Add documentation in `docs/BACKENDS.md` for your backend.

## Pull Request Process

### Before Submitting

1. Ensure tests pass: `make test`
2. Run linters: `make lint`
3. Format code: `make fmt`
4. Update documentation if needed
5. Add/update tests for new functionality

### PR Guidelines

- **Title**: Use clear, descriptive titles
  - Good: "Add HashiCorp Vault backend support"
  - Bad: "Update code"

- **Description**: Include:
  - What changes were made
  - Why the changes were needed
  - How to test the changes
  - Related issue numbers

- **Size**: Keep PRs focused and reasonably sized
  - Large features should be broken into multiple PRs
  - Each PR should be reviewable in < 30 minutes

### Example PR Description

```markdown
## Summary
Adds support for HashiCorp Vault as a key management backend.

## Changes
- Created new `vaultbackend` package
- Implemented SignerBackend interface
- Added configuration parsing for Vault
- Added integration tests with Vault mock

## Testing
```bash
make test
# Or test manually with Vault:
vault server -dev
make run-dev
```

## Related Issues
Closes #42
```

### Review Process

1. Automated checks must pass
2. At least one maintainer review required
3. Address review feedback
4. Maintainer will merge when approved

## Testing

### Unit Tests

```go
func TestMyFunction(t *testing.T) {
    result := MyFunction()
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Tests

For backend integrations, use test containers or mocks:

```go
func TestVaultBackend(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Start test Vault instance
    // Run tests
}
```

Run integration tests:

```bash
go test -v ./...  # includes integration tests
go test -v -short ./...  # skip integration tests
```

## Documentation

### Code Documentation

- Add godoc comments to exported functions and types
- Include examples where helpful

```go
// Sign signs a message using the BLS private key.
// The message is hashed internally before signing.
// Returns the BLS signature bytes or an error.
func (s *BlsSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
    // implementation
}
```

### User Documentation

Update relevant docs in `docs/`:
- `QUICKSTART.md` - Getting started guide
- `PRODUCTION.md` - Production deployment
- `BACKENDS.md` - Backend-specific documentation

## Security

### Reporting Security Issues

**DO NOT** open public issues for security vulnerabilities.

Email security@example.com with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Security Best Practices

- Never log sensitive data (keys, secrets)
- Always validate inputs
- Use constant-time comparisons for secrets
- Follow secure coding guidelines
- Keep dependencies updated

## Release Process

(For maintainers)

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release branch: `release/vX.Y.Z`
4. Test thoroughly
5. Tag release: `git tag vX.Y.Z`
6. Push tag: `git push origin vX.Y.Z`
7. Create GitHub release with notes

## Getting Help

- GitHub Issues: For bugs and feature requests
- GitHub Discussions: For questions and discussions
- Discord: [link] (for real-time chat)

## License

By contributing, you agree that your contributions will be licensed under the BSD-3-Clause License.

## Recognition

Contributors will be added to the CONTRIBUTORS.md file.

---

Thank you for contributing to make Avalanche validators more secure! 🎉
