package server

import (
	"context"
	"fmt"
	"net"

	"github.com/avax-remote-signer/signer/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// MultiServer runs both HTTP and gRPC servers
type MultiServer struct {
	httpServer *Server
	grpcServer *grpc.Server
	grpcAddr   string
	logger     *zap.Logger
}

// MultiServerConfig holds configuration for both protocols
type MultiServerConfig struct {
	HTTPAddress string
	GRPCAddress string
	TLSCert     string
	TLSKey      string
	Logger      *zap.Logger
}

// NewMultiServer creates a server that supports both HTTP and gRPC
func NewMultiServer(blsSigner *signer.BlsSigner, config *MultiServerConfig) (*MultiServer, error) {
	if blsSigner == nil {
		return nil, fmt.Errorf("signer cannot be nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	// Create HTTP server
	httpConfig := &Config{
		Address: config.HTTPAddress,
		TLSCert: config.TLSCert,
		TLSKey:  config.TLSKey,
		Logger:  logger,
	}
	httpServer, err := NewServer(blsSigner, httpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	RegisterGRPCServer(grpcServer, blsSigner)
	
	// Enable reflection for debugging (can be disabled in production)
	reflection.Register(grpcServer)

	return &MultiServer{
		httpServer: httpServer,
		grpcServer: grpcServer,
		grpcAddr:   config.GRPCAddress,
		logger:     logger,
	}, nil
}

// Start starts both HTTP and gRPC servers
func (m *MultiServer) Start() error {
	// Start gRPC server in background
	go func() {
		if err := m.startGRPC(); err != nil {
			m.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// Start HTTP server (blocks)
	httpConfig := &Config{
		Address: m.httpServer.httpServer.Addr,
		Logger:  m.logger,
	}
	return m.httpServer.Start(httpConfig)
}

// startGRPC starts the gRPC server
func (m *MultiServer) startGRPC() error {
	listener, err := net.Listen("tcp", m.grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", m.grpcAddr, err)
	}

	m.logger.Info("Starting gRPC server", zap.String("address", m.grpcAddr))

	if err := m.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}

// Stop gracefully stops both servers
func (m *MultiServer) Stop() error {
	m.logger.Info("Stopping servers")

	// Stop gRPC server
	m.grpcServer.GracefulStop()

	// Stop HTTP server
	if err := m.httpServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	return nil
}

// StopContext stops both servers with a context deadline
func (m *MultiServer) StopContext(ctx context.Context) error {
	m.logger.Info("Stopping servers with context")

	// Stop gRPC server
	stopped := make(chan struct{})
	go func() {
		m.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or context timeout
	select {
	case <-stopped:
		m.logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		m.logger.Warn("gRPC server stop timeout, forcing stop")
		m.grpcServer.Stop()
	}

	// Stop HTTP server
	if err := m.httpServer.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	return nil
}
