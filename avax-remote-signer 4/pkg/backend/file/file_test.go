package filebackend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
)

func TestFileBackend(t *testing.T) {
	// Create a temporary directory for test keys
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")

	// Generate a test key
	if err := GenerateKeyFile(keyPath); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create backend
	backend, err := NewFileBackend(keyPath)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer backend.Close()

	ctx := context.Background()

	// Test GetPublicKey
	pubKey, err := backend.GetPublicKey(ctx)
	if err != nil {
		t.Fatalf("Failed to get public key: %v", err)
	}
	if pubKey == nil {
		t.Fatal("Public key is nil")
	}
	pubKeyBytes := bls.PublicKeyToCompressedBytes(pubKey)
	if len(pubKeyBytes) != bls.PublicKeyLen {
		t.Fatalf("Invalid public key size: got %d, want %d", len(pubKeyBytes), bls.PublicKeyLen)
	}

	// Test Sign
	message := []byte("test message for signing")
	signature, err := backend.Sign(ctx, message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}
	if signature == nil {
		t.Fatal("Signature is nil")
	}
	sigBytes := bls.SignatureToBytes(signature)
	if len(sigBytes) != bls.SignatureLen {
		t.Fatalf("Invalid signature size: got %d, want %d", len(sigBytes), bls.SignatureLen)
	}

	// Verify signature
	if !bls.Verify(pubKey, signature, message) {
		t.Fatal("Signature verification failed")
	}

	t.Logf("Successfully signed and verified message with BLS")
}

func TestFileBackendInvalidPath(t *testing.T) {
	_, err := NewFileBackend("/nonexistent/path/key.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent key file")
	}
}

func TestGenerateKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "generated-key.json")

	if err := GenerateKeyFile(keyPath); err != nil {
		t.Fatalf("Failed to generate key file: %v", err)
	}

	// Check file exists and has correct permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Key file not created: %v", err)
	}

	mode := info.Mode().Perm()
	// avalanchego uses 0400 (read-only) for security
	if mode != 0400 {
		t.Errorf("Expected permissions 0400, got %o", mode)
	}
}
