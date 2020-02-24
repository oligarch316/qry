package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fauxLiteralSuite struct{ decodeSuite }

func newFauxLiteralSuite(level qry.DecodeLevel) (res fauxLiteralSuite) {
	res.level = level
	res.decodeOpts = []qry.Option{
		qry.SetLevelVia(level, qry.SetAllowLiteral),
	}
	return
}

func (fls fauxLiteralSuite) run(t *testing.T) {
	fls.runByteSubtests(t)
	fls.runRuneSubtests(t)
}

func (fls fauxLiteralSuite) runByteSubtests(t *testing.T) {
	var (
		input    = "abc%20三"
		expected = [7]byte{
			0x61,             // a
			0x62,             // b
			0x63,             // c
			0x20,             // <space>
			0xE4, 0xB8, 0x89, // 三 (sān)
		}
	)

	fls.runSubtest(t, "[]byte target", func(t *testing.T, decode tDecode) {
		var target []byte
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected[:], target)
	})

	fls.runSubtest(t, "[7]byte target", func(t *testing.T, decode tDecode) {
		var target [7]byte
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (fls fauxLiteralSuite) runRuneSubtests(t *testing.T) {
	var (
		input    = "abc%20三"
		expected = [5]rune{
			'\u0061', // a
			'\u0062', // b
			'\u0063', // c
			'\u0020', // <space>
			'\u4e09', // 三 (sān)
		}
	)

	fls.runSubtest(t, "[]rune target", func(t *testing.T, decode tDecode) {
		var target []rune
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected[:], target)
	})

	fls.runSubtest(t, "[5]rune target", func(t *testing.T, decode tDecode) {
		var target [5]rune
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}
