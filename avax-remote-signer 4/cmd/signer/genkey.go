package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	filebackend "github.com/avax-remote-signer/signer/pkg/backend/file"
)

var (
	outputPath string
	genKeyCmd  = &cobra.Command{
		Use:   "genkey",
		Short: "Generate a new BLS key pair",
		Long: `Generate a new BLS key pair for Avalanche validator signing.

WARNING: This generates a key file on disk. In production, use a proper
key management system like HashiCorp Vault or a cloud KMS provider.`,
		RunE: generateKey,
	}
)

func init() {
	rootCmd.AddCommand(genKeyCmd)
	genKeyCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path (required)")
	genKeyCmd.MarkFlagRequired("output")
}

func generateKey(cmd *cobra.Command, args []string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("file already exists: %s (will not overwrite)", outputPath)
	}

	fmt.Printf("Generating new BLS key pair...\n")

	if err := filebackend.GenerateKeyFile(outputPath); err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	fmt.Printf("✓ Key pair generated successfully\n")
	fmt.Printf("  Location: %s\n", outputPath)
	fmt.Printf("\n")
	fmt.Printf("IMPORTANT SECURITY NOTES:\n")
	fmt.Printf("  • This file contains your validator private key\n")
	fmt.Printf("  • Store it securely and back it up\n")
	fmt.Printf("  • Set appropriate file permissions (600)\n")
	fmt.Printf("  • Never share this file or commit it to version control\n")
	fmt.Printf("  • For production, use HashiCorp Vault or cloud KMS instead\n")

	return nil
}
