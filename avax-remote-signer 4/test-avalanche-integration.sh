#!/bin/bash
# Test script to simulate Avalanche validator using remote signer

set -e

SIGNER_URL="${SIGNER_URL:-http://localhost:9090}"

echo "=== Avalanche Remote Signer Integration Test ==="
echo ""
echo "Signer URL: $SIGNER_URL"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo -e "${YELLOW}[1/5] Testing Health Endpoint...${NC}"
HEALTH=$(curl -s "$SIGNER_URL/health")
if echo "$HEALTH" | jq -e '.healthy == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Signer is healthy${NC}"
else
    echo -e "${RED}✗ Signer health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Get Validator Public Key
echo -e "${YELLOW}[2/5] Retrieving Validator Public Key...${NC}"
PK_RESPONSE=$(curl -s "$SIGNER_URL/public-key")
PUBLIC_KEY=$(echo "$PK_RESPONSE" | jq -r '.public_key')

if [ -n "$PUBLIC_KEY" ] && [ "$PUBLIC_KEY" != "null" ]; then
    echo -e "${GREEN}✓ Public key retrieved${NC}"
    echo "  Public Key: $PUBLIC_KEY"
    echo "  Length: ${#PUBLIC_KEY} hex chars ($(( ${#PUBLIC_KEY} / 2 )) bytes)"
    
    # Verify BLS format
    if [ ${#PUBLIC_KEY} -eq 96 ]; then
        echo -e "${GREEN}✓ Valid BLS12-381 public key format (48 bytes)${NC}"
    else
        echo -e "${RED}✗ Invalid public key length${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to retrieve public key${NC}"
    exit 1
fi
echo ""

# Test 3: Sign Validator Registration Message
echo -e "${YELLOW}[3/5] Signing Validator Registration Message...${NC}"
# Simulate a validator registration message (NodeID + BLS key proof)
REGISTRATION_MSG="76616c696461746f722d726567697374726174696f6e2d746573742d6d657373616765"
echo "  Message: Validator registration (hex)"

SIGN_RESPONSE=$(curl -s -X POST "$SIGNER_URL/sign" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"$REGISTRATION_MSG\"}")

SIGNATURE=$(echo "$SIGN_RESPONSE" | jq -r '.signature')

if [ -n "$SIGNATURE" ] && [ "$SIGNATURE" != "null" ]; then
    echo -e "${GREEN}✓ Message signed successfully${NC}"
    echo "  Signature: ${SIGNATURE:0:64}..."
    echo "  Length: ${#SIGNATURE} hex chars ($(( ${#SIGNATURE} / 2 )) bytes)"
    
    # Verify BLS signature format
    if [ ${#SIGNATURE} -eq 192 ]; then
        echo -e "${GREEN}✓ Valid BLS12-381 signature format (96 bytes)${NC}"
    else
        echo -e "${RED}✗ Invalid signature length${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to sign message${NC}"
    exit 1
fi
echo ""

# Test 4: Sign Block Proposal
echo -e "${YELLOW}[4/5] Signing Block Proposal...${NC}"
BLOCK_MSG="626c6f636b2d70726f706f73616c2d686173682d74657374"
echo "  Message: Block proposal hash"

BLOCK_SIGN=$(curl -s -X POST "$SIGNER_URL/sign" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"$BLOCK_MSG\"}")

BLOCK_SIG=$(echo "$BLOCK_SIGN" | jq -r '.signature')

if [ -n "$BLOCK_SIG" ] && [ "$BLOCK_SIG" != "null" ]; then
    echo -e "${GREEN}✓ Block proposal signed${NC}"
    echo "  Signature: ${BLOCK_SIG:0:64}..."
else
    echo -e "${RED}✗ Failed to sign block proposal${NC}"
    exit 1
fi
echo ""

# Test 5: Sign Vote
echo -e "${YELLOW}[5/5] Signing Vote Message...${NC}"
VOTE_MSG="766f74652d6d6573736167652d74657374"
echo "  Message: Consensus vote"

VOTE_SIGN=$(curl -s -X POST "$SIGNER_URL/sign" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"$VOTE_MSG\"}")

VOTE_SIG=$(echo "$VOTE_SIGN" | jq -r '.signature')

if [ -n "$VOTE_SIG" ] && [ "$VOTE_SIG" != "null" ]; then
    echo -e "${GREEN}✓ Vote message signed${NC}"
    echo "  Signature: ${VOTE_SIG:0:64}..."
else
    echo -e "${RED}✗ Failed to sign vote${NC}"
    exit 1
fi
echo ""

# Summary
echo "========================================"
echo -e "${GREEN}✓ All Integration Tests Passed!${NC}"
echo "========================================"
echo ""
echo "Summary:"
echo "  • Remote signer is operational"
echo "  • BLS public key: 48 bytes (correct)"
echo "  • BLS signatures: 96 bytes (correct)"
echo "  • All signing operations successful"
echo ""
echo "The remote signer is ready for Avalanche validator integration!"
echo ""
echo "Next Steps:"
echo "  1. Configure avalanchego with --signer-url=$SIGNER_URL"
echo "  2. Provide the BLS public key during validator registration"
echo "  3. The validator will use this signer for all signature operations"
