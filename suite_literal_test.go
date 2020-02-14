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
		qry.SetLevelVia(level, qry.SetAllowLiteral),
	}
	return
}

func (ls literalSuite) run(t *testing.T) {
	ls.runStringSubtest(t)
	ls.runBoolSubtest(t)
	ls.runIntLikeSubtests(t)
	ls.runFloatLikeSubtests(t)
}

func (ls literalSuite) runStringSubtest(t *testing.T) {
	ls.runSubtest(t, "string target", func(t *testing.T, decode tDecode) {
		var (
			input    = "abc%20xyz"
			expected = "abc xyz"
			target   string
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ls literalSuite) runBoolSubtest(t *testing.T) {
	ls.runSubtest(t, "bool target", func(t *testing.T, decode tDecode) {
		var (
			input    = "true"
			expected = true
			target   bool
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ls literalSuite) runIntLikeSubtests(t *testing.T) {
	var (
		input    = "33"
		expected = 33
	)

	ls.runSubtest(t, "int target", func(t *testing.T, decode tDecode) {
		var target int
		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})

	if !testing.Short() {
		ls.runSubtest(t, "int8 target", func(t *testing.T, decode tDecode) {
			var target int8
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int8(expected), target)
		})

		ls.runSubtest(t, "int16 target", func(t *testing.T, decode tDecode) {
			var target int16
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int16(expected), target)
		})

		ls.runSubtest(t, "int32 target", func(t *testing.T, decode tDecode) {
			var target int32
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int32(expected), target)
		})

		ls.runSubtest(t, "int64 target", func(t *testing.T, decode tDecode) {
			var target int64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, int64(expected), target)
		})
	}

	ls.runSubtest(t, "uint target", func(t *testing.T, decode tDecode) {
		var target uint
		require.NoError(t, decode(input, &target))
		assert.Equal(t, uint(expected), target)
	})

	if !testing.Short() {
		ls.runSubtest(t, "uint8 target", func(t *testing.T, decode tDecode) {
			var target uint8
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint8(expected), target)
		})

		ls.runSubtest(t, "uint16 target", func(t *testing.T, decode tDecode) {
			var target uint16
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint16(expected), target)
		})

		ls.runSubtest(t, "uint32 target", func(t *testing.T, decode tDecode) {
			var target uint32
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint32(expected), target)
		})

		ls.runSubtest(t, "uint64 target", func(t *testing.T, decode tDecode) {
			var target uint64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, uint64(expected), target)
		})
	}
}

func (ls literalSuite) runFloatLikeSubtests(t *testing.T) {
	var (
		input    = "2.718"
		expected = 2.718
	)

	ls.runSubtest(t, "float32 target", func(t *testing.T, decode tDecode) {
		var target float32
		require.NoError(t, decode(input, &target))
		assert.Equal(t, float32(expected), target)
	})

	if !testing.Short() {
		ls.runSubtest(t, "float64 target", func(t *testing.T, decode tDecode) {
			var target float64
			require.NoError(t, decode(input, &target))
			assert.Equal(t, float64(expected), target)
		})
	}

	ls.runSubtest(t, "complex64 target", func(t *testing.T, decode tDecode) {
		var target complex64
		require.NoError(t, decode(input, &target))
		assert.Equal(t, complex(float32(expected), 0), target)
	})

	if !testing.Short() {
		ls.runSubtest(t, "complex128 target", func(t *testing.T, decode tDecode) {
			var target complex128
			require.NoError(t, decode(input, &target))
			assert.Equal(t, complex(float64(expected), 0), target)
		})
	}
}
