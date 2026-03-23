package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"go.uber.org/zap"

	"github.com/avax-remote-signer/signer/pkg/signer"
)

// Server implements the HTTP RemoteSigner service
type Server struct {
	signer     *signer.BlsSigner
	logger     *zap.Logger
	httpServer *http.Server
}

// SignRequest represents a signing request
type SignRequest struct {
	Message string `json:"message"` // hex-encoded message
	KeyID   string `json:"key_id,omitempty"`
}

// SignResponse represents a signing response
type SignResponse struct {
	Signature string `json:"signature"` // hex-encoded signature
}

// PublicKeyResponse represents a public key response
type PublicKeyResponse struct {
	PublicKey string `json:"public_key"` // hex-encoded public key
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// Config holds server configuration
type Config struct {
	Address string
	TLSCert string
	TLSKey  string
	Logger  *zap.Logger
}

// NewServer creates a new remote signer gRPC server
func NewServer(blsSigner *signer.BlsSigner, config *Config) (*Server, error) {
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

	server := &Server{
		signer: blsSigner,
		logger: logger,
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/sign", server.handleSign)
	mux.HandleFunc("/public-key", server.handleGetPublicKey)
	mux.HandleFunc("/health", server.handleHealth)

	server.httpServer = &http.Server{
		Addr:    config.Address,
		Handler: mux,
	}

	if config.TLSCert != "" && config.TLSKey != "" {
		logger.Info("TLS enabled for HTTP server")
	} else {
		logger.Warn("TLS not configured - using insecure connection (not recommended for production)")
	}

	return server, nil
}

// handleSign handles signing requests
func (s *Server) handleSign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	messageBytes, err := hex.DecodeString(req.Message)
	if err != nil {
		http.Error(w, "Invalid hex message", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Received sign request", zap.Int("message_size", len(messageBytes)))

	signature, err := s.signer.SignMessage(r.Context(), messageBytes)
	if err != nil {
		s.logger.Error("Failed to sign message", zap.Error(err))
		http.Error(w, "Signing failed", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Successfully signed message")

	resp := SignResponse{
		Signature: hex.EncodeToString(bls.SignatureToBytes(signature)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGetPublicKey handles public key requests
func (s *Server) handleGetPublicKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.logger.Debug("Received get public key request")

	publicKey, err := s.signer.GetPublicKey(r.Context())
	if err != nil {
		s.logger.Error("Failed to get public key", zap.Error(err))
		http.Error(w, "Failed to get public key", http.StatusInternalServerError)
		return
	}

	resp := PublicKeyResponse{
		PublicKey: hex.EncodeToString(bls.PublicKeyToCompressedBytes(publicKey)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	_, err := s.signer.GetPublicKey(r.Context())
	
	resp := HealthResponse{
		Healthy: err == nil,
	}
	if err != nil {
		resp.Message = fmt.Sprintf("Signer unhealthy: %v", err)
	} else {
		resp.Message = "Signer operational"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Start begins serving the HTTP server
func (s *Server) Start(config *Config) error {
	s.logger.Info("Starting HTTP server", zap.String("address", config.Address))

	if config.TLSCert != "" && config.TLSKey != "" {
		if err := s.httpServer.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("TLS server error: %w", err)
		}
	} else {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop() error {
	s.logger.Info("Stopping HTTP server")
	return s.httpServer.Shutdown(context.Background())
}
