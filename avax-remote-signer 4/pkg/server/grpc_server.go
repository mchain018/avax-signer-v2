package server

import (
	"context"

	pb "github.com/avax-remote-signer/signer/proto/rpcdb"
	"github.com/avax-remote-signer/signer/pkg/signer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
)

// GRPCServer implements the gRPC Signer service
type GRPCServer struct {
	pb.UnimplementedSignerServer
	signer *signer.BlsSigner
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(s *signer.BlsSigner) *GRPCServer {
	return &GRPCServer{
		signer: s,
	}
}

// Sign implements the Sign RPC method
func (s *GRPCServer) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	if len(req.Message) == 0 {
		return nil, status.Error(codes.InvalidArgument, "message cannot be empty")
	}

	// Sign the message
	signature, err := s.signer.SignMessage(ctx, req.Message)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to sign message: %v", err)
	}

	// Convert signature to bytes
	sigBytes := bls.SignatureToBytes(signature)

	return &pb.SignResponse{
		Signature: sigBytes,
	}, nil
}

// PublicKey implements the PublicKey RPC method
func (s *GRPCServer) PublicKey(ctx context.Context, req *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
	// Get public key from signer
	pubKey, err := s.signer.GetPublicKey(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get public key: %v", err)
	}

	// Convert public key to compressed bytes (48 bytes for BLS12-381)
	pkBytes := bls.PublicKeyToCompressedBytes(pubKey)

	return &pb.PublicKeyResponse{
		PublicKey: pkBytes,
	}, nil
}

// RegisterGRPCServer registers the gRPC server with the given gRPC server instance
func RegisterGRPCServer(grpcServer *grpc.Server, signer *signer.BlsSigner) {
	pb.RegisterSignerServer(grpcServer, NewGRPCServer(signer))
}

// SignProofOfPossession implements the SignProofOfPossession RPC method
// This MUST use the BLS_POP_ domain separation tag, not BLS_SIG_
func (s *GRPCServer) SignProofOfPossession(ctx context.Context, req *pb.SignProofOfPossessionRequest) (*pb.SignProofOfPossessionResponse, error) {
	// Sign proof of possession using the BLS_POP_ DST
	// This is different from regular signing and is required for validator registration
	signature, err := s.signer.SignProofOfPossession(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to sign proof of possession: %v", err)
	}

	// Convert signature to bytes
	sigBytes := bls.SignatureToBytes(signature)

	return &pb.SignProofOfPossessionResponse{
		Signature: sigBytes,
	}, nil
}
