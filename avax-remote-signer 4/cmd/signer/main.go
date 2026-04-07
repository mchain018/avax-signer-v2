package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	filebackend "github.com/avax-remote-signer/signer/pkg/backend/file"
	"github.com/avax-remote-signer/signer/pkg/config"
	"github.com/avax-remote-signer/signer/pkg/server"
	"github.com/avax-remote-signer/signer/pkg/signer"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "avax-remote-signer",
		Short: "Remote signing service for Avalanche validators",
		Long: `A remote signer sidecar for Avalanche validators that provides BLS signature
operations while keeping private keys isolated from the validator node.

Supports pluggable backends for key management including file-based storage,
HashiCorp Vault, and cloud KMS providers.`,
		RunE: runServer,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger
	logger, err := setupLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting Avalanche Remote Signer",
		zap.String("backend", cfg.Backend.Type),
		zap.String("grpc_address", cfg.Server.GRPCAddress),
		zap.String("http_address", cfg.Server.HTTPAddress),
	)

	// Initialize backend
	backend, err := createBackend(cfg.Backend, logger)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}
	defer backend.Close()

	// Create BLS signer
	blsSigner := signer.NewBlsSigner(backend)

	// Log the public key
	ctx := context.Background()
	pubKey, err := blsSigner.GetPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	logger.Info("Validator public key loaded",
		zap.String("public_key", fmt.Sprintf("%x", pubKey.Serialize())),
	)

	// Create multi-protocol server (HTTP + gRPC)
	serverCfg := &server.MultiServerConfig{
		HTTPAddress: cfg.Server.HTTPAddress,
		GRPCAddress: cfg.Server.GRPCAddress,
		Logger:      logger,
	}
	if cfg.Server.TLS.Enabled {
		serverCfg.TLSCert = cfg.Server.TLS.CertFile
		serverCfg.TLSKey = cfg.Server.TLS.KeyFile
	}

	multiServer, err := server.NewMultiServer(blsSigner, serverCfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		multiServer.Stop()
	}()

	// Start both servers (blocks until stopped)
	if err := multiServer.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info("Server stopped")
	return nil
}

func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var zapConfig zap.Config
	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure file output if specified
	if cfg.File != "" {
		// Ensure log directory exists
		logDir := filepath.Dir(cfg.File)
		if logDir != "" && logDir != "." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}
		}

		// Add file output paths
		zapConfig.OutputPaths = []string{
			"stdout",           // Log to stdout as well
			cfg.File,           // Log to file
		}
		zapConfig.ErrorOutputPaths = []string{
			"stderr",           // Error goes to stderr
			cfg.File,           // Error goes to file too
		}
	}

	return zapConfig.Build()
}

func createBackend(cfg config.BackendConfig, logger *zap.Logger) (signer.SignerBackend, error) {
	switch cfg.Type {
	case "file":
		keyPath, ok := cfg.Config["key_path"].(string)
		if !ok || keyPath == "" {
			return nil, fmt.Errorf("file backend requires 'key_path' in config")
		}
		logger.Warn("Using file-based backend - NOT RECOMMENDED FOR PRODUCTION")
		return filebackend.NewFileBackend(keyPath)

	// Add more backend types here as they're implemented
	// case "vault":
	//     return vaultbackend.New(cfg.Config, logger)
	// case "awskms":
	//     return awsbackend.New(cfg.Config, logger)

	default:
		return nil, fmt.Errorf("unsupported backend type: %s", cfg.Type)
	}
}
