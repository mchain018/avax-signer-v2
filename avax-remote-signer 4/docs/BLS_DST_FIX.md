# BLS Domain Separation Tag Fix

## Issue

AvalancheGo uses two different BLS domain separation tags (DST) for signing operations:

- **Regular signing**: `BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_`
- **Proof of Possession**: `BLS_POP_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_`

The difference is subtle but critical: `BLS_SIG_` vs `BLS_POP_`.

If a signer uses the regular signing DST for both operations, the output will be a structurally valid BLS signature but will fail PoP verification every time.

## Fix

The signer now properly implements both signing methods:

1. `Sign()` - Uses the `BLS_SIG_` domain separation tag for regular message signing
2. `SignProofOfPossession()` - Uses the `BLS_POP_` domain separation tag for signing the compressed public key bytes

## Implementation Details

### Stack Changes

1. **SignerBackend interface** (`pkg/signer/backend.go`):
   - Added `SignProofOfPossession(ctx context.Context) (*bls.Signature, error)` method

2. **FileBackend** (`pkg/backend/file/file.go`):
   - Implemented `SignProofOfPossession()` that:
     - Gets the public key bytes (compressed, 48 bytes)
     - Calls `localsigner.SignProofOfPossession(pkBytes)` which uses the `BLS_POP_` DST

3. **BlsSigner** (`pkg/signer/bls.go`):
   - Added `SignProofOfPossession()` method that delegates to the backend

4. **gRPC Server** (`pkg/server/grpc_server.go`):
   - Fixed `SignProofOfPossession` RPC handler to call `signer.SignProofOfPossession()` instead of `signer.SignMessage()`

## Verification Steps

To verify the fix works correctly:

1. **Test with local key file**:
   ```bash
   avalanchego --staking-signer-key-file=/path/to/key
   curl -X POST --data '{"jsonrpc":"2.0","id":1,"method":"info.getNodeID"}' \
     -H 'content-type:application/json;' http://127.0.0.1:9650/ext/info
   ```
   Save the `nodeID` and `pop` from the response.

2. **Test with remote signer**:
   ```bash
   # Start the remote signer
   ./signer -config config.yaml

   # Start avalanchego with remote signer
   avalanchego --staking-rpc-signer-endpoint=localhost:50051

   # Get node ID and PoP
   curl -X POST --data '{"jsonrpc":"2.0","id":1,"method":"info.getNodeID"}' \
     -H 'content-type:application/json;' http://127.0.0.1:9650/ext/info
   ```

3. **Compare results**:
   - Public key (`nodeID`) should be identical
   - Proof of possession (`pop`) should now also be identical
   - If they match, the fix is working correctly!

## References

From the AvalancheGo repository:

- Ciphersuite definitions: `utils/crypto/bls/ciphersuite.go`
- PoP generation and verification: `vms/platformvm/signer/proof_of_possession.go`
- Local signer reference: `utils/crypto/bls/signer/localsigner/localsigner.go`
- gRPC proto: `proto/signer/signer.proto`
- RPC signer client: `utils/crypto/bls/signer/rpcsigner/client.go`

## Technical Details

The proof of possession is created by signing the compressed public key bytes (48 bytes for BLS12-381 G1 points) using the `BLS_POP_` ciphersuite. This is a different operation from regular message signing and requires a different domain separation tag to prevent signature reuse attacks and ensure proper security properties.

The `localsigner.SignProofOfPossession(publicKeyBytes)` method in AvalancheGo handles this automatically by using the correct ciphersuite internally.
