package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/require"
)

// ===== Suite
type unmarshalerSuite struct{ decodeSuite }

func newUnmarshalerSuite(level qry.DecodeLevel) (res unmarshalerSuite) {
	res.level = level
	return
}

func (us unmarshalerSuite) run(t *testing.T) {
	var (
		rawText       = "abc%20xyz"
		unescapedText = "abc xyz"
	)

	us.runSubtest(t, "*tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tUnmarshaler
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, unescapedText)
	})

	us.runSubtest(t, "**tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tUnmarshaler)
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, unescapedText)
	})

	us.runSubtest(t, "*tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tRawUnmarshaler
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, rawText)
	})

	us.runSubtest(t, "**tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tRawUnmarshaler)
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, rawText)
	})
}
