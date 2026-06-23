package admin

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

/*
@file abi_helper.go

@desc
ABI encoding utilities for preparing unsigned Ethereum transaction calldata.
Used by admin services to encode function calls for governance execution.

@responsibilities
- Compute 4-byte function selectors via Keccak256
- Encode address and uint256 arguments into 32-byte ABI slots
- Build full calldata hex strings for TransactionPreparation responses
- Extract context.Context from the ctx interface{} passed by handlers
*/

/*
@function extractCtx

@desc
Extracts a context.Context from the interface{} value passed by admin handlers.
Handlers pass either *gin.Context or context.Context; this function handles both.

@params
- ctx: handler-supplied context (either *gin.Context or context.Context)

@returns
- context.Context: a valid context, falling back to context.Background()
*/
func extractCtx(ctx interface{}) context.Context {
	if gc, ok := ctx.(*gin.Context); ok {
		return gc.Request.Context()
	}
	if c, ok := ctx.(context.Context); ok {
		return c
	}
	return context.Background()
}

/*
@function functionSelector

@desc
Computes the 4-byte ABI function selector for the given Solidity function signature.

@params
- sig: full Solidity signature, e.g. "setMinCollateralRatio(address,uint256)"

@returns
- []byte: 4-byte selector
*/
func functionSelector(sig string) []byte {
	return crypto.Keccak256([]byte(sig))[:4]
}

/*
@function padAddress

@desc
ABI-encodes an Ethereum address as a 32-byte slot (12 zero bytes + 20 address bytes).

@params
- addr: hex address string with or without 0x prefix

@returns
- []byte: 32-byte ABI-encoded address slot
*/
func padAddress(addr string) []byte {
	addr = strings.TrimPrefix(strings.ToLower(addr), "0x")
	// left-pad to 20 bytes
	if len(addr) > 40 {
		addr = addr[len(addr)-40:]
	}
	decoded, _ := hex.DecodeString(addr)
	slot := make([]byte, 32)
	copy(slot[32-len(decoded):], decoded)
	return slot
}

/*
@function padUint256

@desc
ABI-encodes a big.Int as a 32-byte big-endian slot.

@params
- val: integer value (may be nil, treated as zero)

@returns
- []byte: 32-byte ABI-encoded uint256 slot
*/
func padUint256(val *big.Int) []byte {
	slot := make([]byte, 32)
	if val == nil {
		return slot
	}
	b := val.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(slot[32-len(b):], b)
	return slot
}

/*
@function padUint256FromInt

@desc
Convenience wrapper for padUint256 that accepts int64.

@params
- val: integer value

@returns
- []byte: 32-byte ABI-encoded uint256 slot
*/
func padUint256FromInt(val int64) []byte {
	return padUint256(big.NewInt(val))
}

/*
@function buildCalldata

@desc
Concatenates a function selector with its ABI-encoded argument slots.

@params
- selector: 4-byte function selector
- args: zero or more 32-byte argument slots

@returns
- string: 0x-prefixed hex calldata string
*/
func buildCalldata(selector []byte, args ...[]byte) string {
	data := make([]byte, 0, 4+len(args)*32)
	data = append(data, selector...)
	for _, arg := range args {
		data = append(data, arg...)
	}
	return "0x" + hex.EncodeToString(data)
}

/*
@function hexEncodeBytes

@desc
Returns the hex encoding of a byte slice without 0x prefix.
*/
func hexEncodeBytes(b []byte) string {
	return hex.EncodeToString(b)
}

/*
@function padStringBytes

@desc
Pads a UTF-8 string byte slice to the next 32-byte boundary for ABI encoding.

@params
- b: raw string bytes

@returns
- []byte: string bytes padded with trailing zeros to a multiple of 32
*/
func padStringBytes(b []byte) []byte {
	n := len(b)
	if n == 0 {
		return make([]byte, 32)
	}
	padded := ((n + 31) / 32) * 32
	out := make([]byte, padded)
	copy(out, b)
	return out
}
