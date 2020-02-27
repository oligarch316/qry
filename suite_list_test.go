package qry_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Types
type (
	tByteUnmarshaler    byte
	tByteRawUnmarshaler byte

	tRuneUnmarshaler    rune
	tRuneRawUnmarshaler rune
)

func (tbu *tByteUnmarshaler) UnmarshalText(text []byte) error {
	*tbu = tByteUnmarshaler(text[0])
	return nil
}

func (tbru *tByteRawUnmarshaler) UnmarshalRawText(text []byte) error {
	*tbru = tByteRawUnmarshaler(text[0])
	return nil
}

func (tru *tRuneUnmarshaler) UnmarshalText(text []byte) error {
	r, _ := utf8.DecodeRune(text)
	*tru = tRuneUnmarshaler(r)
	return nil
}

func (trru *tRuneRawUnmarshaler) UnmarshalRawText(text []byte) error {
	r, _ := utf8.DecodeRune(text)
	*trru = tRuneRawUnmarshaler(r)
	return nil
}

// ===== Error
func (des decodeErrorSuite) runListTests(t *testing.T, separator string) {
	des.runSubtest(t, "array too small error", func(t *testing.T, decode tDecode) {
		var (
			input  = strings.Join([]string{"item A", "item B", "item C"}, separator)
			target [2]string
		)

		actual := decode(input, &target)
		assertErrorMessage(t, "insufficient destination array length", actual)
	})
}

// ===== Success
func (dss decodeSuccessSuite) runListTests(t *testing.T, separator string) {
	t.Run("replace", func(t *testing.T) { dss.runListReplaceSubtests(t, separator) })
	t.Run("update", func(t *testing.T) { dss.runListUpdateSubtests(t, separator) })
	t.Run("pseudo faux literal", func(t *testing.T) { dss.runListPseudoFauxLiteralSubtests(t, separator) })
}

func (dss decodeSuccessSuite) runListReplaceSubtests(t *testing.T, separator string) {
	var (
		rawItems       = []string{"item%20A", "item%20B", "item%20C"}
		unescapedItems = [3]string{"item A", "item B", "item C"}

		inputText = strings.Join(rawItems, separator)
		runner    = dss.withSetOpts(qry.SetReplaceContainer)
	)

	runner.runSubtest(t, "non-zero []string target", func(t *testing.T, decode tDecode) {
		target := []string{"oldOne", "oldTwo"}
		decode(inputText, &target)
		assert.Equal(t, unescapedItems[:], target)
	})

	runner.runSubtest(t, "zero []string target", func(t *testing.T, decode tDecode) {
		var target []string
		decode(inputText, &target)
		assert.Equal(t, unescapedItems[:], target)
	})

	runner.runSubtest(t, "[3]string target", func(t *testing.T, decode tDecode) {
		var target [3]string
		decode(inputText, &target)
		assert.Equal(t, unescapedItems, target)
	})
}

func (dss decodeSuccessSuite) runListUpdateSubtests(t *testing.T, separator string) {
	var (
		rawItems       = []string{"item%20A", "item%20B", "item%20C"}
		unescapedItems = [3]string{"item A", "item B", "item C"}

		inputText = strings.Join(rawItems, separator)
		runner    = dss.withSetOpts(qry.SetUpdateContainer)
	)

	runner.runSubtest(t, "non-zero []string target", func(t *testing.T, decode tDecode) {
		var (
			original = []string{"origOne", "origTwo"}
			target   = make([]string, len(original))
			expected = append(original, unescapedItems[:]...)
		)

		copy(target, original)

		decode(inputText, &target)
		assert.Equal(t, expected, target)
	})

	runner.runSubtest(t, "zero []string target", func(t *testing.T, decode tDecode) {
		var target []string
		decode(inputText, &target)
		assert.Equal(t, unescapedItems[:], target)
	})

	runner.runSubtest(t, "[3]string target", func(t *testing.T, decode tDecode) {
		var target [3]string
		decode(inputText, &target)
		assert.Equal(t, unescapedItems, target)
	})
}

func (dss decodeSuccessSuite) runListPseudoFauxLiteralSubtests(t *testing.T, separator string) {
	var (
		rawItems  = []string{"%20%20", "三三", "四四"}
		inputText = strings.Join(rawItems, separator)
	)

	dss.runSubtest(t, "[]tByteUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <space> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x20, 0xE4, 0xE5}
			target   []tByteUnmarshaler
		)

		decode(inputText, &target)
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[]tByteRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <percent> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x25, 0xE4, 0xE5}
			target   []tByteRawUnmarshaler
		)

		decode(inputText, &target)
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[]tRuneUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <space>, <sān>, <sì>
			expected = []rune{'\u0020', '\u4E09', '\u56DB'}
			target   []tRuneUnmarshaler
		)

		decode(inputText, &target)
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[]tRuneRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <percent>, <sān>, <sì>
			expected = []rune{'\u0025', '\u4E09', '\u56DB'}
			target   []tRuneRawUnmarshaler
		)

		decode(inputText, &target)
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[3]tByteUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <space> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x20, 0xE4, 0xE5}
			target   [3]tByteUnmarshaler
		)

		decode(inputText, &target)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[3]tByteRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <percent> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x25, 0xE4, 0xE5}
			target   [3]tByteRawUnmarshaler
		)

		decode(inputText, &target)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[3]tRuneUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <space>, <sān>, <sì>
			expected = []rune{'\u0020', '\u4E09', '\u56DB'}
			target   [3]tRuneUnmarshaler
		)

		decode(inputText, &target)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	dss.runSubtest(t, "[3]tRuneRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <percent>, <sān>, <sì>
			expected = []rune{'\u0025', '\u4E09', '\u56DB'}
			target   [3]tRuneRawUnmarshaler
		)

		decode(inputText, &target)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})
}
