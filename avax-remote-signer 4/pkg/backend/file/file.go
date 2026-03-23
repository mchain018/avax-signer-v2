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
	keyPath string
	signer  *localsigner.LocalSigner
	mu      sync.RWMutex
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

	return fb, nil
}

// loadKey reads and parses the BLS key from the file
func (fb *FileBackend) loadKey() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	// Use avalanchego's localsigner to load from file
	signer, err := localsigner.FromFile(fb.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load key from file: %w", err)
	}

	fb.signer = signer.(*localsigner.LocalSigner)
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
func (fb *FileBackend) SignProofOfPossession(ctx context.Context) (*bls.Signature, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.signer == nil {
		return nil, fmt.Errorf("signer not loaded")
	}

	// Get the public key bytes (compressed, 48 bytes)
	pubKey := fb.signer.PublicKey()
	pkBytes := bls.PublicKeyToCompressedBytes(pubKey)

	// Sign proof of possession using BLS_POP_ DST
	// This signs the public key bytes with a different domain separation tag
	sig, err := fb.signer.SignProofOfPossession(pkBytes)
	if err != nil {
		return nil, fmt.Errorf("signing proof of possession failed: %w", err)
	}

	return sig, nil
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
