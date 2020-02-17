package qry_test

import (
	"strings"
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listSuite struct {
	replaceMode, updateMode decodeSuite
	separator               string
}

func newListSuite(level qry.DecodeLevel, separator string) listSuite {
	ds := decodeSuite{
		level:      level,
		decodeOpts: []qry.Option{qry.SetAllLevelsVia(qry.SetAllowLiteral)},
	}

	return listSuite{
		replaceMode: ds.with(qry.SetLevelVia(level, qry.SetReplaceContainer)),
		updateMode:  ds.with(qry.SetLevelVia(level, qry.SetUpdateContainer)),
		separator:   separator,
	}
}

func (ls listSuite) run(t *testing.T) {
	t.Run("replace", ls.runReplaceSubtests)
	t.Run("update", ls.runUpdateSubtests)
	t.Run("pseudo faux literal", ls.runPseudoFauxLiteralSubtests)
}

func (ls listSuite) runReplaceSubtests(t *testing.T) {
	var (
		rawItems       = []string{"item%20A", "item%20B", "item%20C"}
		unescapedItems = [3]string{"item A", "item B", "item C"}

		inputText = strings.Join(rawItems, ls.separator)
	)

	ls.replaceMode.runSubtest(t, "non-zero []string target", func(t *testing.T, decode tDecode) {
		target := []string{"oldOne", "oldTwo"}
		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, unescapedItems[:], target)
	})

	ls.replaceMode.runSubtest(t, "zero []string target", func(t *testing.T, decode tDecode) {
		var target []string
		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, unescapedItems[:], target)
	})

	ls.replaceMode.runSubtest(t, "[3]string target", func(t *testing.T, decode tDecode) {
		var target [3]string
		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, unescapedItems, target)
	})
}

func (ls listSuite) runUpdateSubtests(t *testing.T) {
	var (
		rawItems       = []string{"item%20A", "item%20B", "item%20C"}
		unescapedItems = [3]string{"item A", "item B", "item C"}

		inputText = strings.Join(rawItems, ls.separator)
	)

	ls.updateMode.runSubtest(t, "non-zero []string target", func(t *testing.T, decode tDecode) {
		var (
			original = []string{"origOne", "origTwo"}
			target   = make([]string, len(original))
			expected = append(original, unescapedItems[:]...)
		)

		copy(target, original)

		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, expected, target)
	})

	ls.updateMode.runSubtest(t, "zero []string target", func(t *testing.T, decode tDecode) {
		var target []string
		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, unescapedItems[:], target)
	})

	ls.replaceMode.runSubtest(t, "[3]string target", func(t *testing.T, decode tDecode) {
		var target [3]string
		require.NoError(t, decode(inputText, &target))
		assert.Equal(t, unescapedItems, target)
	})
}

func (ls listSuite) runPseudoFauxLiteralSubtests(t *testing.T) {
	var (
		rawItems  = []string{"%20%20", "三三", "四四"}
		inputText = strings.Join(rawItems, ls.separator)
	)

	ls.replaceMode.runSubtest(t, "[]tByteUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <space> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x20, 0xE4, 0xE5}
			target   []tByteUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[]tByteRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <percent> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x25, 0xE4, 0xE5}
			target   []tByteRawUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[]tRuneUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <space>, <sān>, <sì>
			expected = []rune{'\u0020', '\u4E09', '\u56DB'}
			target   []tRuneUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[]tRuneRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <percent>, <sān>, <sì>
			expected = []rune{'\u0025', '\u4E09', '\u56DB'}
			target   []tRuneRawUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		require.Len(t, target, 3)
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[3]tByteUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <space> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x20, 0xE4, 0xE5}
			target   [3]tByteUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[3]tByteRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// <percent> byte, <sān> leading byte, <sì> leading byte
			expected = []byte{0x25, 0xE4, 0xE5}
			target   [3]tByteRawUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[3]tRuneUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <space>, <sān>, <sì>
			expected = []rune{'\u0020', '\u4E09', '\u56DB'}
			target   [3]tRuneUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})

	ls.replaceMode.runSubtest(t, "[3]tRuneRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			// utf8 code points for <percent>, <sān>, <sì>
			expected = []rune{'\u0025', '\u4E09', '\u56DB'}
			target   [3]tRuneRawUnmarshaler
		)

		require.NoError(t, decode(inputText, &target))
		for i, item := range target {
			assert.EqualValuesf(t, expected[i], item, "index %d", i)
		}
	})
}
