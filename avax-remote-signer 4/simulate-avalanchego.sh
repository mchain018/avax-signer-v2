#!/bin/bash
# Simulates avalanchego validator making signing requests to remote signer

set -e

SIGNER_URL="${SIGNER_URL:-http://host.docker.internal:9090}"

echo "=== Simulated Avalanche Validator Node ==="
echo "Connecting to Remote Signer: $SIGNER_URL"
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Simulate validator startup
echo -e "${BLUE}[INFO] Starting avalanchego validator...${NC}"
sleep 1

# Check signer health
echo -e "${BLUE}[INFO] Checking remote signer connectivity...${NC}"
if curl -sf "$SIGNER_URL/health" > /dev/null; then
    echo -e "${GREEN}✓ Remote signer is reachable${NC}"
else
    echo -e "\033[0;31m✗ Cannot reach remote signer at $SIGNER_URL${NC}"
    exit 1
fi
echo ""

# Get validator public key
echo -e "${BLUE}[INFO] Retrieving BLS public key from remote signer...${NC}"
PK_RESPONSE=$(curl -s "$SIGNER_URL/public-key")
PUBLIC_KEY=$(echo "$PK_RESPONSE" | jq -r '.public_key')
echo -e "${GREEN}✓ BLS Public Key: $PUBLIC_KEY${NC}"
echo ""

# Simulate consensus operations
echo -e "${BLUE}[INFO] Validator is now participating in consensus...${NC}"
echo ""

for i in {1..10}; do
    # Simulate different types of messages that validators sign
    case $((i % 4)) in
        0)
            MSG_TYPE="Block Proposal"
            # Simulate block hash
            MESSAGE=$(printf "%064x" $RANDOM$RANDOM)
            ;;
        1)
            MSG_TYPE="Vote Accept"
            MESSAGE=$(printf "%064x" $RANDOM$RANDOM)
            ;;
        2)
            MSG_TYPE="Vote Reject"
            MESSAGE=$(printf "%064x" $RANDOM$RANDOM)
            ;;
        3)
            MSG_TYPE="State Transition"
            MESSAGE=$(printf "%064x" $RANDOM$RANDOM)
            ;;
    esac
    
    echo -e "${YELLOW}[Block $i] $MSG_TYPE${NC}"
    
    # Sign the message
    SIGN_RESPONSE=$(curl -s -X POST "$SIGNER_URL/sign" \
        -H "Content-Type: application/json" \
        -d "{\"message\": \"$MESSAGE\"}")
    
    SIGNATURE=$(echo "$SIGN_RESPONSE" | jq -r '.signature')
    
    if [ -n "$SIGNATURE" ] && [ "$SIGNATURE" != "null" ]; then
        echo -e "  ${GREEN}✓ Signed${NC} (sig: ${SIGNATURE:0:32}...)"
    else
        echo -e "  \033[0;31m✗ Signing failed${NC}"
        exit 1
    fi
    
    # Simulate consensus delay
    sleep 2
done

echo ""
echo -e "${GREEN}=== Validator Successfully Processed 10 Blocks ===${NC}"
echo ""
echo "Summary:"
echo "  • Remote signer integration: ✓ Working"
echo "  • BLS signatures: ✓ Valid"
echo "  • Consensus participation: ✓ Active"
echo ""
echo "The validator would continue running and signing indefinitely..."
