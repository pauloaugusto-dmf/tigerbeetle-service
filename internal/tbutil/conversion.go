package tbutil

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Uint128ToUint64Safe safely converts a Uint128 to a uint64.
// Returns an error if the high 64 bits are not zero.
func Uint128ToUint64Safe(u types.Uint128) (uint64, error) {
	for i := 8; i < 16; i++ {
		if u[i] != 0 {
			return 0, errors.New("value exceeds uint64 capacity")
		}
	}
	return binary.LittleEndian.Uint64(u[0:8]), nil
}

// Uint128ToString converts a little-endian Uint128 to a decimal string.
func Uint128ToString(u types.Uint128) string {
	reversed := make([]byte, 16)
	copy(reversed, u[:])

	// Reverse bytes (little-endian to big-endian)
	for i := 0; i < 8; i++ {
		reversed[i], reversed[15-i] = u[15-i], u[i]
	}

	return new(big.Int).SetBytes(reversed).String()
}

// ParseUint128FromString parses a decimal string to a little-endian Uint128.
// Returns an error if the value is negative or exceeds 128 bits.
func ParseUint128FromString(s string) (types.Uint128, error) {
	i := new(big.Int)
	if _, ok := i.SetString(s, 10); !ok {
		return types.Uint128{}, errors.New("failed to convert string to big.Int")
	}
	if i.Sign() < 0 {
		return types.Uint128{}, errors.New("negative values are not supported for Uint128")
	}

	b := i.Bytes()
	if len(b) > 16 {
		return types.Uint128{}, errors.New("value exceeds 128-bit limit")
	}

	var u types.Uint128
	copy(u[16-len(b):], b)

	// Convert to little-endian
	for i := 0; i < 8; i++ {
		u[i], u[15-i] = u[15-i], u[i]
	}

	return u, nil
}
