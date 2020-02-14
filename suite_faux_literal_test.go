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
		// Ensure "allow literal" is true for the targeted level
		qry.SetLevelVia(level, qry.SetAllowLiteral),
	}
	return
}

func (fls fauxLiteralSuite) subtests(t *testing.T) {
	fls.byteSubtests(t)
	fls.runeSubtests(t)
}

func (fls fauxLiteralSuite) byteSubtests(t *testing.T) {
	var (
		input    = "abc%20xyz"
		expected = [7]byte{0x61, 0x62, 0x63, 0x20, 0x78, 0x79, 0x7A}
	)

	fls.run(t, "[]byte target", func(t *testing.T, decode tDecode) {
		var target []byte
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected[:], target)
	})

	fls.run(t, "[7]byte target", func(t *testing.T, decode tDecode) {
		var target [7]byte
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (fls fauxLiteralSuite) runeSubtests(t *testing.T) {
	var (
		input    = "abc%20xyz"
		expected = [7]rune{'a', 'b', 'c', ' ', 'x', 'y', 'z'}
	)

	fls.run(t, "[]rune target", func(t *testing.T, decode tDecode) {
		var target []rune
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected[:], target)
	})

	fls.run(t, "[7]rune target", func(t *testing.T, decode tDecode) {
		var target [7]rune
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}
