package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

type (
	testByteUnmarshaler    byte
	testByteRawUnmarshaler byte
)

func (tbu *testByteUnmarshaler) UnmarshalText([]byte) error {
	*tbu = testByteUnmarshaler(1)
	return nil
}
func (tbru *testByteRawUnmarshaler) UnmarshalRawText([]byte) error {
	*tbru = testByteRawUnmarshaler(1)
	return nil
}

func (tbu testByteUnmarshaler) assert(t *testing.T) bool { return assert.Equal(t, byte(1), byte(tbu)) }
func (tbru testByteRawUnmarshaler) assert(t *testing.T) bool {
	return assert.Equal(t, byte(1), byte(tbru))
}

type (
	testRuneUnmarshaler    rune
	testRuneRawUnmarshaler rune
)

func (tru *testRuneUnmarshaler) UnmarshalText([]byte) error {
	*tru = testRuneUnmarshaler(1)
	return nil
}
func (trru *testRuneRawUnmarshaler) UnmarshalRawText([]byte) error {
	*trru = testRuneRawUnmarshaler(1)
	return nil
}

func (tru testRuneUnmarshaler) assert(t *testing.T) bool { return assert.Equal(t, rune(1), rune(tru)) }
func (trru testRuneRawUnmarshaler) assert(t *testing.T) bool {
	return assert.Equal(t, rune(1), rune(trru))
}

func testFauxLiteralErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("unescape", func(t *testing.T) {
		t.Skip("TODO: converter.Unescape error")
	})

	t.Run("array too small", func(t *testing.T) {
		t.Skip("TODO: insufficient target length error")
	})
}

func testFauxLiterals(t *testing.T, level qry.DecodeLevel) {
	var (
		rawInput       = "abc%20xyz"
		unescapedInput = "abc xyz"

		base = newTest(
			configOptionsAs(qry.SetLevelVia(level, qry.SetAllowLiteral)),
			decodeLevelAs(level),
			inputAs(rawInput),
		)
	)

	t.Run("[]byte target", func(t *testing.T) {
		var target []byte

		trace := base.require(t, &target)
		if !assert.Equal(t, unescapedInput, string(target)) {
			trace.log(t)
		}
	})

	t.Run("[]rune target", func(t *testing.T) {
		var target []rune

		trace := base.require(t, &target)
		if !assert.Equal(t, unescapedInput, string(target)) {
			trace.log(t)
		}
	})

	t.Run("[7]byte target", func(t *testing.T) {
		var (
			target   [7]byte
			expected = [7]byte{'a', 'b', 'c', ' ', 'x', 'y', 'z'}
		)

		trace := base.require(t, &target)
		if !assert.Equal(t, expected, target) {
			trace.log(t)
		}
	})

	t.Run("[7]rune target", func(t *testing.T) {
		var (
			target   [7]rune
			expected = [7]rune{'a', 'b', 'c', ' ', 'x', 'y', 'z'}
		)

		trace := base.require(t, &target)
		if !assert.Equal(t, expected, target) {
			trace.log(t)
		}
	})
}

// Big TODO:
// func testFauxUnmarshalers(t *testing.T, level qry.DecodeLevel) {
//
//     t.Run("[]ByteUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]ByteRawUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]RuneUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]RuneRawUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]ByteUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]ByteRawUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]RuneUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]RuneRawUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]*ByteUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]*ByteRawUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]*RuneUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run("[]*RuneRawUnmarshaler target", func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]*ByteUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]*ByteRawUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]*RuneUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
//
//     t.Run(fmt.Sprintf("[%d]*RuneRawUnmarshaler target", len(unescapedInput)), func(t *testing.T) {
//         t.Skip("TODO")
//     })
// }
