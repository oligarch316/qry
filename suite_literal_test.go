package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type literalSuite struct{ decodeSuite }

func newLiteralSuite(level qry.DecodeLevel) (res literalSuite) {
	res.level = level
	res.decodeOpts = []qry.Option{
		// Ensure "allow literal" is true for the targeted level
		qry.SetLevelVia(level, qry.SetAllowLiteral),
	}
	return
}

func (ls literalSuite) subtests(t *testing.T) {
	ls.stringSubtest(t)
	ls.boolSubtest(t)
	ls.intLikeSubtests(t)
	ls.floatLikeSubtests(t)
}

func (ls literalSuite) stringSubtest(t *testing.T) {
	ls.run(t, "string target", func(t *testing.T, decode tDecode) {
		var (
			input    = "abc%20xyz"
			expected = "abc xyz"
			target   string
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ls literalSuite) boolSubtest(t *testing.T) {
	ls.run(t, "bool target", func(t *testing.T, decode tDecode) {
		var (
			input    = "true"
			expected = true
			target   bool
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ls literalSuite) intLikeSubtests(t *testing.T) {
	var (
		input    = "33"
		expected = 33
	)

	ls.run(t, "int target", func(t *testing.T, decode tDecode) {
		var target int
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})

	if !testing.Short() {
		ls.run(t, "int8 target", func(t *testing.T, decode tDecode) {
			var target int8
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int8(expected), target)
		})

		ls.run(t, "int16 target", func(t *testing.T, decode tDecode) {
			var target int16
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int16(expected), target)
		})

		ls.run(t, "int32 target", func(t *testing.T, decode tDecode) {
			var target int32
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int32(expected), target)
		})

		ls.run(t, "int64 target", func(t *testing.T, decode tDecode) {
			var target int64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int64(expected), target)
		})
	}

	ls.run(t, "uint target", func(t *testing.T, decode tDecode) {
		var target uint
		require.NoError(t, decode(input, &target))
		assert.Equal(t, uint(expected), target)
	})

	if !testing.Short() {
		ls.run(t, "uint8 target", func(t *testing.T, decode tDecode) {
			var target uint8
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint8(expected), target)
		})

		ls.run(t, "uint16 target", func(t *testing.T, decode tDecode) {
			var target uint16
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint16(expected), target)
		})

		ls.run(t, "uint32 target", func(t *testing.T, decode tDecode) {
			var target uint32
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint32(expected), target)
		})

		ls.run(t, "uint64 target", func(t *testing.T, decode tDecode) {
			var target uint64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint64(expected), target)
		})
	}
}

func (ls literalSuite) floatLikeSubtests(t *testing.T) {
	var (
		input    = "2.718"
		expected = 2.718
	)

	ls.run(t, "float32 target", func(t *testing.T, decode tDecode) {
		var target float32
		require.NoError(t, decode(input, &target))
		assert.Equal(t, float32(expected), target)
	})

	if !testing.Short() {
		ls.run(t, "float64 target", func(t *testing.T, decode tDecode) {
			var target float64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, float64(expected), target)
		})
	}

	ls.run(t, "complex64 target", func(t *testing.T, decode tDecode) {
		var target complex64
		require.NoError(t, decode(input, &target))
		assert.Equal(t, complex(float32(expected), 0), target)
	})

	if !testing.Short() {
		ls.run(t, "complex128 target", func(t *testing.T, decode tDecode) {
			var target complex128
			require.NoError(t, decode(input, &target))
			assert.Equal(t, complex(float64(expected), 0), target)
		})
	}
}
