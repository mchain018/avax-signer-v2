package filebackend

import (
	"context"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
)

// FileBackend implements SignerBackend using a local file for key storage
// WARNING: This is for DEVELOPMENT/TESTING only. Do not use in production!
type FileBackend struct {
	keyPath    string
	signer     *localsigner.LocalSigner
	publicKey  *bls.PublicKey
	pkBytes    []byte
	popCache   *bls.Signature // Cache PoP to ensure it never changes
	mu         sync.RWMutex
}

// NewFileBackend creates a new file-based signer backend
func NewFileBackend(keyPath string) (*FileBackend, error) {
	if keyPath == "" {
		return nil, fmt.Errorf("key path cannot be empty")
	}

	fb := &FileBackend{
		keyPath: keyPath,
	}

	// Load the key on initialization
	if err := fb.loadKey(); err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	}

	// Pre-compute and cache the Proof of Possession once
	if err := fb.computeAndCachePOP(); err != nil {
		return nil, fmt.Errorf("failed to compute PoP: %w", err)
	}

	return fb, nil
}

// loadKey reads and parses the BLS key from the file (only on init)
func (fb *FileBackend) loadKey() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	// Use avalanchego's localsigner to load from file
	signer, err := localsigner.FromFile(fb.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load key from file: %w", err)
	}

	fb.signer = signer.(*localsigner.LocalSigner)
	
	// Cache the public key and its bytes
	fb.publicKey = fb.signer.PublicKey()
	fb.pkBytes = bls.PublicKeyToCompressedBytes(fb.publicKey)
	
	return nil
}

// computeAndCachePOP pre-computes and caches the PoP to ensure determinism
func (fb *FileBackend) computeAndCachePOP() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.signer == nil {
		return fmt.Errorf("signer not loaded")
	}

	// Compute PoP once and cache it
	// This is deterministic as long as we use the same key and public key bytes
	sig, err := fb.signer.SignProofOfPossession(fb.pkBytes)
	if err != nil {
		return fmt.Errorf("failed to compute PoP: %w", err)
	}

	fb.popCache = sig
	return nil
}

// Sign implements SignerBackend.Sign using BLS_SIG_ domain separation tag
func (fb *FileBackend) Sign(ctx context.Context, message []byte) (*bls.Signature, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.signer == nil {
		return nil, fmt.Errorf("signer not loaded")
	}

	// Sign using BLS with BLS_SIG_ DST
	sig, err := fb.signer.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	return sig, nil
}

// SignProofOfPossession implements SignerBackend.SignProofOfPossession using BLS_POP_ domain separation tag
// Returns the cached PoP which is deterministic and never changes for this validator
func (fb *FileBackend) SignProofOfPossession(ctx context.Context) (*bls.Signature, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.signer == nil {
		return nil, fmt.Errorf("signer not loaded")
	}

	// Return cached PoP - this ensures the PoP is always the same for this validator
	// The PoP is computed once during initialization and never changes
	if fb.popCache == nil {
		return nil, fmt.Errorf("proof of possession not initialized")
	}

	return fb.popCache, nil
}

// GetPublicKey implements SignerBackend.GetPublicKey
func (fb *FileBackend) GetPublicKey(ctx context.Context) (*bls.PublicKey, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.signer == nil {
		return nil, fmt.Errorf("signer not loaded")
	}

	return fb.signer.PublicKey(), nil
}

// Close implements SignerBackend.Close
func (fb *FileBackend) Close() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.signer != nil {
		return fb.signer.Shutdown()
	}
	return nil
}

// GenerateKeyFile creates a new BLS key pair and saves it to a file
// This is a utility function for initial setup
func GenerateKeyFile(outputPath string) error {
	// Generate new key pair using avalanchego's localsigner
	signer, err := localsigner.New()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Save to file
	if err := signer.ToFile(outputPath); err != nil {
		return fmt.Errorf("failed to save key to file: %w", err)
	}

	// Cleanup
	return signer.Shutdown()
}
