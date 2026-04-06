package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/crypto/bls/signer/localsigner"
)

var (
	importKeyCmd = &cobra.Command{
		Use:   "importkey",
		Short: "Import a BLS key or generate a new one for validator use",
		Long: `Import an existing BLS key or generate a new BLS12-381 key pair for Avalanche validator signing.

This command creates a binary BLS key file compatible with the remote signer.

NOTE: If you have an existing AvalancheGo validator with a nodeID, you cannot 
directly convert the existing Ed25519 staking key to a BLS key. Instead, you must:
1. Generate a new BLS key pair using this tool
2. Register your validator with the new BLS public key
3. Use this remote signer for all future signing operations`,
		RunE: importKey,
	}
)

var (
	outputKeyPath string
)

func init() {
	rootCmd.AddCommand(importKeyCmd)
	importKeyCmd.Flags().StringVarP(&outputKeyPath, "output", "o", "", "output file path for the BLS key (required)")
	importKeyCmd.MarkFlagRequired("output")
}

func importKey(cmd *cobra.Command, args []string) error {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Avalanche Remote Signer - BLS Key Import/Generation Tool")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Check if file already exists
	if _, err := os.Stat(outputKeyPath); err == nil {
		return fmt.Errorf("file already exists: %s (will not overwrite)", outputKeyPath)
	}

	fmt.Println("IMPORTANT INFORMATION:")
	fmt.Println()
	fmt.Println("1. KEY FORMAT:")
	fmt.Println("   - This tool generates a NEW BLS12-381 key pair")
	fmt.Println("   - If you have an existing Avalanche validator, you CANNOT convert its")
	fmt.Println("     existing Ed25519 staking key to BLS format")
	fmt.Println("   - Each validator needs its own unique BLS key pair")
	fmt.Println()
	fmt.Println("2. WHAT TO DO WITH AN EXISTING VALIDATOR:")
	fmt.Println("   a) Generate a new BLS key using this tool")
	fmt.Println("   b) Register a NEW validator with the new BLS public key")
	fmt.Println("   c) Keep this key file safe and back it up")
	fmt.Println("   d) Configure the remote signer to use this key file")
	fmt.Println()
	fmt.Println("3. MIGRATION FROM EXISTING VALIDATOR:")
	fmt.Println("   If you want to migrate an existing validator to use BLS:")
	fmt.Println("   - You must unregister the old validator")
	fmt.Println("   - Generate a new BLS key pair (this tool)")
	fmt.Println("   - Register a new validator with the new BLS key")
	fmt.Println("   - This may incur a lock-up period for unstaked tokens")
	fmt.Println()
	fmt.Println()

	fmt.Println("Generating new BLS key pair...")
	fmt.Println()

	// Generate new BLS key
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return fmt.Errorf("failed to generate random seed: %w", err)
	}

	// Create private key from seed
	sk, err := bls.SecretKeyFromSeed(seed)
	if err != nil {
		return fmt.Errorf("failed to create secret key: %w", err)
	}

	// Create public key
	pk := bls.PublicFromSecretKey(sk)

	// Create signer from key
	signer := localsigner.NewSigner(sk, pk)

	// Save to file
	if err := signer.ToFile(outputKeyPath); err != nil {
		return fmt.Errorf("failed to save key to file: %w", err)
	}

	// Verify by reading it back
	verifyLoc, err := localsigner.FromFile(outputKeyPath)
	if err != nil {
		return fmt.Errorf("failed to verify saved key: %w", err)
	}

	verifiedSigner := verifyLoc.(*localsigner.LocalSigner)
	pubKeyBytes := bls.PublicKeyToCompressedBytes(verifiedSigner.PublicKey())

	// Cleanup
	signer.Shutdown()
	verifiedSigner.Shutdown()

	// Display results
	fmt.Println("✓ BLS Key Pair Generated Successfully!")
	fmt.Println()
	fmt.Println("KEY DETAILS:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Location: %s\n", outputKeyPath)
	fmt.Printf("Public Key (hex): %x\n", pubKeyBytes)
	fmt.Printf("Public Key Size: %d bytes\n", len(pubKeyBytes))
	fmt.Println()
	fmt.Println("NEXT STEPS:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("1. Secure Storage:")
	fmt.Printf("   chmod 600 %s\n", outputKeyPath)
	fmt.Printf("   Backup: cp %s /secure/backup/\n", outputKeyPath)
	fmt.Println()
	fmt.Println("2. Configure Remote Signer:")
	fmt.Println("   Update your config.yaml:")
	fmt.Printf("   backend:\n")
	fmt.Printf("     type: \"file\"\n")
	fmt.Printf("     config:\n")
	fmt.Printf("       key_path: \"%s\"\n", outputKeyPath)
	fmt.Println()
	fmt.Println("3. Register Validator:")
	fmt.Println("   Use this public key when registering a new validator:")
	fmt.Printf("   %x\n", pubKeyBytes)
	fmt.Println()
	fmt.Println("4. Start Remote Signer:")
	fmt.Println("   ./avax-remote-signer --config config.yaml")
	fmt.Println()
	fmt.Println("SECURITY NOTES:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("✓ File contains unencrypted private key - protect it like an SSH key")
	fmt.Println("✓ Back it up to secure, offline storage")
	fmt.Println("✓ Never commit to version control")
	fmt.Println("✓ For production, consider HashiCorp Vault or cloud KMS")
	fmt.Println()

	return nil
}
