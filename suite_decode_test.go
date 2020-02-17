package qry_test

import (
	"testing"
	"unicode/utf8"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

// ===== Decode test types

// Basic unmarshalers
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

// Unmarshaler that is not a pointer kind
type tNonPointerUnmarshaler func(string)

func (tnpu tNonPointerUnmarshaler) UnmarshalText(text []byte) error {
	tnpu(string(text))
	return nil
}

// Unmarshalers with byte and rune kinds
// > used to test that slices of these are not mistaken for faux literals
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

// ===== Decode test runner

type (
	tDecode       func(string, interface{}) error
	decodeSubtest func(*testing.T, tDecode)
)

type decodeSuite struct {
	level      qry.DecodeLevel
	decodeOpts []qry.Option
}

func (ds decodeSuite) with(opts ...qry.Option) decodeSuite {
	res := decodeSuite{level: ds.level}
	if len(ds.decodeOpts) > 0 {
		res.decodeOpts = make([]qry.Option, len(ds.decodeOpts))
		copy(res.decodeOpts, ds.decodeOpts)
	}
	res.decodeOpts = append(res.decodeOpts, opts...)
	return res
}

func (ds decodeSuite) dumpTraceOnFailure(t *testing.T, trace *qry.TraceTree) {
	if t.Failed() {
		t.Logf("Decode Trace:\n%s\n", trace.Sdump())
	}
}

func (ds decodeSuite) runSubtest(t *testing.T, name string, subtest decodeSubtest) {
	t.Run(name, func(t *testing.T) {
		// Setup
		var (
			decoder = qry.NewDecoder(ds.decodeOpts...)
			trace   = qry.NewTraceTree()
		)

		// Teardown
		if testing.Verbose() {
			// Use defer here to ensure trace dump in FailNow() situations
			defer ds.dumpTraceOnFailure(t, trace)
		}

		// Test
		assert.NotPanics(t, func() {
			subtest(t, func(input string, v interface{}) error {
				return decoder.Decode(ds.level, input, v, trace)
			})
		})
	})
}
