package signer

import (
	"context"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
)

// SignerBackend is the interface that all key management backends must implement
// This allows plugging in different storage solutions (file, Vault, KMS, HSM, etc.)
type SignerBackend interface {
	// Sign signs a message using the BLS_SIG_ domain separation tag
	Sign(ctx context.Context, message []byte) (*bls.Signature, error)

	// SignProofOfPossession signs the public key bytes using the BLS_POP_ domain separation tag
	// This is required for validator registration and uses a different DST than regular signing
	SignProofOfPossession(ctx context.Context) (*bls.Signature, error)

	// GetPublicKey returns the BLS public key for this validator
	GetPublicKey(ctx context.Context) (*bls.PublicKey, error)

	// Close performs cleanup when shutting down
	Close() error
}
