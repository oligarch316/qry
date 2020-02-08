package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

func testLiteralErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("convert", func(t *testing.T) {
		t.Skip("TODO: all possible strconv.XYZ errors")
	})

	t.Run("unescape", func(t *testing.T) {
		t.Skip("TODO: converter.Unescape error")
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Skip("TODO: unmarshaler error")
	})
}

func testLiterals(t *testing.T, level qry.DecodeLevel) {
	base := newTest(
		configOptionsAs(qry.SetLevelVia(level, qry.SetAllowLiteral)),
		decodeLevelAs(level),
	)

	// ----- Text
	var (
		rawText       = "abc%20xyz"
		unescapedText = "abc xyz"
		textTest      = base.with(inputAs(rawText))
	)

	t.Run("string target", func(t *testing.T) {
		var target string

		trace := textTest.require(t, &target)
		if !assert.Equal(t, unescapedText, target) {
			trace.log(t)
		}
	})

	t.Run("CustomString target", func(t *testing.T) {
		var target TestCustomString

		trace := textTest.require(t, &target)
		if !target.assert(t, unescapedText) {
			trace.log(t)
		}
	})

	t.Run("TextUnmarshaler target", func(t *testing.T) {
		var target TestUnmarshaler

		trace := textTest.require(t, &target)
		if !target.assert(t, unescapedText) {
			trace.log(t)
		}
	})

	t.Run("RawTextUnmarshaler target", func(t *testing.T) {
		var target TestRawUnmarshaler

		trace := textTest.require(t, &target)
		if !target.assert(t, rawText) {
			trace.log(t)
		}
	})

	// ----- Boolean
	t.Run("bool target", func(t *testing.T) {
		var target bool

		trace := base.with(inputAs("true")).require(t, &target)
		if !assert.True(t, target) {
			trace.log(t)
		}
	})

	// ----- Integer
	intTest := base.with(inputAs("42"))

	t.Run("int target", func(t *testing.T) {
		var target int

		trace := intTest.require(t, &target)
		if !assert.Equal(t, 42, target) {
			trace.log(t)
		}
	})

	t.Run("int8 target", func(t *testing.T) {
		var target int8

		trace := intTest.require(t, &target)
		if !assert.Equal(t, int8(42), target) {
			trace.log(t)
		}
	})

	t.Run("int16 target", func(t *testing.T) {
		var target int16

		trace := intTest.require(t, &target)
		if !assert.Equal(t, int16(42), target) {
			trace.log(t)
		}
	})

	t.Run("int32 target", func(t *testing.T) {
		var target int32

		trace := intTest.require(t, &target)
		if !assert.Equal(t, int32(42), target) {
			trace.log(t)
		}
	})

	t.Run("int64 target", func(t *testing.T) {
		var target int64

		trace := intTest.require(t, &target)
		if !assert.Equal(t, int64(42), target) {
			trace.log(t)
		}
	})

	// TODO:
	// - all uint type
	// - all complex types
}
