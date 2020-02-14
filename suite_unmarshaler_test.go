package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tUnmarshalerData struct {
	called bool
	val    string
}

func (tud *tUnmarshalerData) doUnmarshal(text []byte) error {
	tud.called = true
	tud.val = string(text)
	return nil
}

func (tud tUnmarshalerData) assertCalledWithValue(t *testing.T, val string) bool {
	return assert.True(t, tud.called, "check unmarshal function was called") && assert.Equal(t, val, tud.val)
}

type (
	tUnmarshaler    struct{ tUnmarshalerData }
	tRawUnmarshaler struct{ tUnmarshalerData }
)

func (tu *tUnmarshaler) UnmarshalText(text []byte) error        { return tu.doUnmarshal(text) }
func (tru *tRawUnmarshaler) UnmarshalRawText(text []byte) error { return tru.doUnmarshal(text) }

type unmarshalerSuite struct{ decodeSuite }

func newUnmarshalerSuite(level qry.DecodeLevel) (res unmarshalerSuite) {
	res.level = level
	return
}

func (us unmarshalerSuite) subtests(t *testing.T) {
	var (
		rawText       = "abc%20xyz"
		unescapedText = "abc xyz"
	)

	us.run(t, "*tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tUnmarshaler
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, unescapedText)
	})

	us.run(t, "**tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tUnmarshaler)
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, unescapedText)
	})

	us.run(t, "*tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tRawUnmarshaler
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, rawText)
	})

	us.run(t, "**tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tRawUnmarshaler)
		require.NoError(t, decode(rawText, &target))
		target.assertCalledWithValue(t, rawText)
	})
}
