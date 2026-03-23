#!/bin/bash
# Test gRPC integration with the remote signer using grpcurl

set -e

GRPC_URL="${GRPC_URL:-localhost:9090}"

echo "=== Avalanche Remote Signer gRPC Integration Test ==="
echo ""
echo "gRPC Server: $GRPC_URL"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${YELLOW}Installing grpcurl...${NC}"
    brew install grpcurl
fi

echo -e "${BLUE}[1/3] Testing gRPC Reflection${NC}"
if grpcurl -plaintext $GRPC_URL list 2>/dev/null | grep -q "signer.Signer"; then
    echo -e "${GREEN}✓ gRPC service discovered${NC}"
else
    echo -e "${RED}✗ gRPC service not available${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}[2/3] Calling PublicKey RPC${NC}"
PUBLIC_KEY_RESPONSE=$(grpcurl -plaintext -d '{}' $GRPC_URL signer.Signer/PublicKey 2>/dev/null)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ PublicKey RPC succeeded${NC}"
    echo "$PUBLIC_KEY_RESPONSE" | jq -r '.publicKey' | xxd -p -r | xxd -p -c 48
else
    echo -e "${RED}✗ PublicKey RPC failed${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}[3/3] Calling Sign RPC${NC}"
# Create a test message (hex: "hello world")
TEST_MSG="68656c6c6f20776f726c64"
SIGN_REQUEST=$(echo -n "$TEST_MSG" | xxd -p -r | base64)

SIGN_RESPONSE=$(grpcurl -plaintext -d "{\"message\": \"$SIGN_REQUEST\"}" $GRPC_URL signer.Signer/Sign 2>/dev/null)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Sign RPC succeeded${NC}"
    echo "Message: hello world"
    echo -n "Signature: "
    echo "$SIGN_RESPONSE" | jq -r '.signature' | base64 -d | xxd -p -c 96
else
    echo -e "${RED}✗ Sign RPC failed${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}[4/4] Calling SignProofOfPossession RPC${NC}"
POP_RESPONSE=$(grpcurl -plaintext -d '{}' $GRPC_URL signer.Signer/SignProofOfPossession 2>/dev/null)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ SignProofOfPossession RPC succeeded${NC}"
    echo -n "Proof: "
    echo "$POP_RESPONSE" | jq -r '.signature' | base64 -d | xxd -p -c 96
else
    echo -e "${RED}✗ SignProofOfPossession RPC failed${NC}"
    exit 1
fi
echo ""

echo "========================================"
echo -e "${GREEN}✓ All gRPC Tests Passed!${NC}"
echo "========================================"
echo ""
echo "Summary:"
echo "  • gRPC service is operational"
echo "  • PublicKey RPC: ✓ Working"
echo "  • Sign RPC: ✓ Working"
echo "  • SignProofOfPossession RPC: ✓ Working"
echo ""
echo "The remote signer is ready for avalanchego integration!"
