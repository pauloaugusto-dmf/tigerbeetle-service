package tbutil_test

import (
	"math/big"
	"testing"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/tbutil"
	"github.com/stretchr/testify/assert"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func TestUint128ToUint64Safe(t *testing.T) {
	t.Run("valid uint64 conversion", func(t *testing.T) {
		var u types.Uint128
		copy(u[:8], []byte{0x78, 0x56, 0x34, 0x12}) // little endian of 0x12345678
		val, err := tbutil.Uint128ToUint64Safe(u)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0x12345678), val)
	})

	t.Run("value exceeds uint64", func(t *testing.T) {
		var u types.Uint128
		u[15] = 1 // High byte not zero
		_, err := tbutil.Uint128ToUint64Safe(u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value exceeds")
	})
}

func TestUint128ToStringAndParse(t *testing.T) {
	testCases := []string{
		"0",
		"1",
		"18446744073709551615", // max uint64
		"340282366920938463463374607431768211455", // max uint128
		"123456789012345678901234567890",
	}

	for _, input := range testCases {
		t.Run("round-trip "+input, func(t *testing.T) {
			parsed, err := tbutil.ParseUint128FromString(input)
			assert.NoError(t, err)

			output := tbutil.Uint128ToString(parsed)

			iInput := new(big.Int)
			iOutput := new(big.Int)
			iInput.SetString(input, 10)
			iOutput.SetString(output, 10)

			assert.Equal(t, 0, iInput.Cmp(iOutput))
		})
	}

	t.Run("negative input", func(t *testing.T) {
		_, err := tbutil.ParseUint128FromString("-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "negative")
	})

	t.Run("value exceeds 128 bits", func(t *testing.T) {
		tooBig := new(big.Int).Exp(big.NewInt(2), big.NewInt(130), nil).String()
		_, err := tbutil.ParseUint128FromString(tooBig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds 128")
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := tbutil.ParseUint128FromString("not_a_number")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert")
	})
}
