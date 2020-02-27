package qry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===== Error
func (des decodeErrorSuite) runFauxLiteralTests(t *testing.T) {
	des.runFauxLiteralUnescapeSubtests(t)
	des.runFauxLiteralArraySizeSubtests(t)
}

func (des decodeErrorSuite) runFauxLiteralUnescapeSubtests(t *testing.T) {
	des.withUnescapeError("forced unescape error").runSubtest(t, "unescape error", func(t *testing.T, decode tDecode) {
		var target []rune
		actual := decode("xyz", &target)
		assertErrorMessage(t, "forced unescape error", actual)
	})
}

func (des decodeErrorSuite) runFauxLiteralArraySizeSubtests(t *testing.T) {
	var (
		/*
		 * a       | 1 byte, 1 rune
		 * b       | 1 byte, 1 rune
		 * c       | 1 byte, 1 rune
		 * <space> | 1 byte, 1 rune
		 * 三 (sān) | 3 bytes, 1 rune
		 *
		 * Total   | 7 bytes, 5 runes
		 */

		input    = "abc%20三"
		expected = "insufficient destination array length"
	)

	des.runSubtest(t, "byte array too small error", func(t *testing.T, decode tDecode) {
		var target [6]byte
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "rune array too small error", func(t *testing.T, decode tDecode) {
		var target [4]rune
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})
}

// ===== Success
func (dss decodeSuccessSuite) runFauxLiteralTests(t *testing.T) {
	dss.runFauxLiteralByteSubtests(t)
	dss.runFauxLiteralRuneSubtests(t)
}

func (dss decodeSuccessSuite) runFauxLiteralByteSubtests(t *testing.T) {
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

	dss.runSubtest(t, "[]byte target", func(t *testing.T, decode tDecode) {
		var target []byte
		decode(input, &target)
		assert.Equal(t, expected[:], target)
	})

	dss.runSubtest(t, "[7]byte target", func(t *testing.T, decode tDecode) {
		var target [7]byte
		decode(input, &target)
		assert.Equal(t, expected, target)
	})
}

func (dss decodeSuccessSuite) runFauxLiteralRuneSubtests(t *testing.T) {
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

	dss.runSubtest(t, "[]rune target", func(t *testing.T, decode tDecode) {
		var target []rune
		decode(input, &target)
		assert.Equal(t, expected[:], target)
	})

	dss.runSubtest(t, "[5]rune target", func(t *testing.T, decode tDecode) {
		var target [5]rune
		decode(input, &target)
		assert.Equal(t, expected, target)
	})
}
