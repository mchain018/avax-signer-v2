package signer

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
)

// BlsSigner handles BLS signature operations for Avalanche validators
type BlsSigner struct {
	backend SignerBackend
}

// NewBlsSigner creates a new BLS signer with the specified backend
func NewBlsSigner(backend SignerBackend) *BlsSigner {
	return &BlsSigner{
		backend: backend,
	}
}

// SignMessage signs a message using BLS signature scheme with BLS_SIG_ DST
func (s *BlsSigner) SignMessage(ctx context.Context, message []byte) (*bls.Signature, error) {
	if len(message) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	// Get the signature from the backend using BLS_SIG_ DST
	sig, err := s.backend.Sign(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("backend signing failed: %w", err)
	}

	return sig, nil
}

// SignProofOfPossession signs the proof of possession using BLS_POP_ DST
// This is required for validator registration and must use a different domain separation tag
func (s *BlsSigner) SignProofOfPossession(ctx context.Context) (*bls.Signature, error) {
	// Sign proof of possession using BLS_POP_ DST
	sig, err := s.backend.SignProofOfPossession(ctx)
	if err != nil {
		return nil, fmt.Errorf("backend proof of possession signing failed: %w", err)
	}

	return sig, nil
}

// GetPublicKey retrieves the validator's BLS public key
func (s *BlsSigner) GetPublicKey(ctx context.Context) (*bls.PublicKey, error) {
	pk, err := s.backend.GetPublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	return pk, nil
}
