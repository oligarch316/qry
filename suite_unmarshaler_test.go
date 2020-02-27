package qry_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===== Types

// ----- Unmarshaler that is not a pointer kind (non-pointer receiver)
// > Used in indirect and map success tests
type tNonPointerUnmarshaler func(string)

func (tnpu tNonPointerUnmarshaler) UnmarshalText(text []byte) error {
	tnpu(string(text))
	return nil
}

// ----- Error unmarshalers
type (
	tErrorUnmarshaler    struct{ forcedErr error }
	tErrorRawUnmarshaler struct{ forcedErr error }
)

func (teu tErrorUnmarshaler) UnmarshalText(text []byte) error        { return teu.forcedErr }
func (teru tErrorRawUnmarshaler) UnmarshalRawText(text []byte) error { return teru.forcedErr }

// ----- Basic unmarshalers
type (
	tUnmarshalerData struct {
		called bool
		val    string
	}
	tUnmarshaler    struct{ tUnmarshalerData }
	tRawUnmarshaler struct{ tUnmarshalerData }
)

func (tud *tUnmarshalerData) doUnmarshal(text []byte) error {
	tud.called = true
	tud.val = string(text)
	return nil
}

func (tud tUnmarshalerData) assertCalledWithValue(t *testing.T, val string) bool {
	return assert.True(t, tud.called, "check unmarshal function was called") && assert.Equal(t, val, tud.val)
}

func (tu *tUnmarshaler) UnmarshalText(text []byte) error        { return tu.doUnmarshal(text) }
func (tru *tRawUnmarshaler) UnmarshalRawText(text []byte) error { return tru.doUnmarshal(text) }

// ===== Error
func (des decodeErrorSuite) runUnmarshalerTests(t *testing.T) {
	des.runUnmarshalerUnescapeSubtest(t)
	des.runUnmarshalerResultSubtests(t)
}

func (des decodeErrorSuite) runUnmarshalerUnescapeSubtest(t *testing.T) {
	des.withUnescapeError("forced unescape error").runSubtest(t, "unescape error", func(t *testing.T, decode tDecode) {
		var target tUnmarshaler
		actual := decode("xyz", &target)
		assertErrorMessage(t, "forced unescape error", actual)
	})
}

func (des decodeErrorSuite) runUnmarshalerResultSubtests(t *testing.T) {
	var (
		forcedUnmarshalMsg = "forced unmarshal error"
		forcedUnmarshalErr = errors.New(forcedUnmarshalMsg)
	)

	des.runSubtest(t, "unmarshal text error", func(t *testing.T, decode tDecode) {
		target := tErrorUnmarshaler{forcedUnmarshalErr}
		actual := decode("xyz", &target)
		assertErrorMessage(t, forcedUnmarshalMsg, actual)
	})

	des.runSubtest(t, "unmarshal raw text error", func(t *testing.T, decode tDecode) {
		target := tErrorRawUnmarshaler{forcedUnmarshalErr}
		actual := decode("xyz", &target)
		assertErrorMessage(t, forcedUnmarshalMsg, actual)
	})
}

// ===== Success
func (dss decodeSuccessSuite) runUnmarshalerTests(t *testing.T) {
	var (
		rawText       = "abc%20xyz"
		unescapedText = "abc xyz"
	)

	dss.runSubtest(t, "*tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tUnmarshaler
		decode(rawText, &target)
		target.assertCalledWithValue(t, unescapedText)
	})

	dss.runSubtest(t, "**tUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tUnmarshaler)
		decode(rawText, &target)
		target.assertCalledWithValue(t, unescapedText)
	})

	dss.runSubtest(t, "*tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target tRawUnmarshaler
		decode(rawText, &target)
		target.assertCalledWithValue(t, rawText)
	})

	dss.runSubtest(t, "**tRawUnmarshaler target", func(t *testing.T, decode tDecode) {
		var target = new(tRawUnmarshaler)
		decode(rawText, &target)
		target.assertCalledWithValue(t, rawText)
	})
}
