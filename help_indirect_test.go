package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

func testIndirects(t *testing.T, level qry.DecodeLevel) {
	// TODO!!!: Test update mode for indirects

	var (
		rawInput       = "abc%20xyz"
		unescapedInput = "abc xyz"

		base = newTest(
			configOptionsAs(qry.SetLevelVia(level, qry.SetAllowLiteral)),
			decodeLevelAs(level),
			inputAs(rawInput),
		)
	)

	t.Run("wants literal", func(t *testing.T) {
		t.Run("*string target", func(t *testing.T) {
			var target *string

			trace := base.require(t, &target)
			if !assert.Equal(t, unescapedInput, *target) {
				trace.log(t)
			}
		})

		t.Run("*CustomString target", func(t *testing.T) {
			var target *TestCustomString

			trace := base.require(t, &target)
			if !target.assert(t, unescapedInput) {
				trace.log(t)
			}
		})

		t.Run("*[]byte target", func(t *testing.T) {
			t.Skip("TODO")
		})

		t.Run("*[]rune target", func(t *testing.T) {
			t.Skip("TODO")
		})

		// TODO: Is there need for *[]testByteUnmarshaler, etc... here?

		t.Run("*TextUnmarshaler target", func(t *testing.T) {
			var target *TestUnmarshaler

			trace := base.require(t, &target)
			if !target.assert(t, unescapedInput) {
				trace.log(t)
			}
		})

		t.Run("*RawTextUnmarshaler target", func(t *testing.T) {
			var target *TestRawUnmarshaler

			trace := base.require(t, &target)
			if !target.assert(t, rawInput) {
				trace.log(t)
			}
		})

		t.Run("interface{string} target", func(t *testing.T) {
			var (
				tmp    string
				target interface{} = tmp
			)

			trace := base.require(t, &target)
			if !assert.Equal(t, unescapedInput, target) {
				trace.log(t)
			}
		})

		t.Run("interface{[]byte} target", func(t *testing.T) {
			t.Skip("TODO")
		})

		t.Run("interface{[]rune} target", func(t *testing.T) {
			t.Skip("TODO")
		})

		t.Run("interface{CustomString} target", func(t *testing.T) {
			var (
				tmp    TestCustomString
				target TestStringAsserter = tmp
			)

			trace := base.require(t, &target)
			if !target.assert(t, unescapedInput) {
				trace.log(t)
			}
		})

		t.Run("interface{TextUnmarshaler} target", func(t *testing.T) {
			var (
				tmp    TestUnmarshaler
				target TestStringAsserter = tmp
			)

			trace := base.require(t, &target)
			if !target.assert(t, unescapedInput) {
				trace.log(t)
			}
		})

		t.Run("interface{RawTextUnmarshaler} target", func(t *testing.T) {
			var (
				tmp    TestRawUnmarshaler
				target TestStringAsserter = tmp
			)

			trace := base.require(t, &target)
			if !target.assert(t, rawInput) {
				trace.log(t)
			}
		})
	})

	t.Run("has literal", func(t *testing.T) {
		t.Run("*string target", func(t *testing.T) {
			var (
				target         string
				indirectTarget = &target
			)

			trace := base.require(t, &indirectTarget)
			if !(assert.Equal(t, unescapedInput, *indirectTarget) && assert.Equal(t, unescapedInput, target)) {
				trace.log(t)
			}
		})

		t.Run("*[]byte target", func(t *testing.T) {
			t.Skip("TODO")
		})

		t.Run("*[]rune target", func(t *testing.T) {
			t.Skip("TODO")
		})

		t.Run("*CustomString target", func(t *testing.T) {
			var (
				target         TestCustomString
				indirectTarget = &target
			)

			trace := base.require(t, &indirectTarget)
			if !(indirectTarget.assert(t, unescapedInput) && target.assert(t, unescapedInput)) {
				trace.log(t)
			}
		})

		t.Run("*TextUnmarshaler target", func(t *testing.T) {
			var (
				target         TestUnmarshaler
				indirectTarget = &target
			)

			trace := base.require(t, &indirectTarget)
			if !(indirectTarget.assert(t, unescapedInput) && target.assert(t, unescapedInput)) {
				trace.log(t)
			}
		})

		t.Run("*RawTextUnmarshaler target", func(t *testing.T) {
			var (
				target         TestRawUnmarshaler
				indirectTarget = &target
			)

			trace := base.require(t, &indirectTarget)
			if !(indirectTarget.assert(t, rawInput) && target.assert(t, rawInput)) {
				trace.log(t)
			}
		})
	})

	t.Run("has indirect that wants literal", func(t *testing.T) {
		t.Run("*interface{string} target", func(t *testing.T) {
			var (
				tmp         string
				indirectOne interface{} = tmp
				indirectTwo             = &indirectOne
			)

			trace := base.require(t, &indirectTwo)
			if !(assert.Equal(t, unescapedInput, *indirectTwo) && assert.Equal(t, unescapedInput, indirectOne)) {
				trace.log(t)
			}
		})

		t.Run("interface{*string} target", func(t *testing.T) {
			var (
				tmp    *string
				target interface{} = tmp
			)

			trace := base.require(t, &target)
			if !assert.Equal(t, unescapedInput, *(target.(*string))) {
				trace.log(t)
			}
		})
	})

	t.Run("has indirect that has literal", func(t *testing.T) {
		t.Run("interface{*string} target", func(t *testing.T) {
			var (
				target      string
				indirectOne             = &target
				indirectTwo interface{} = indirectOne
			)

			trace := base.require(t, &indirectTwo)
			success :=
				assert.Equal(t, unescapedInput, *(indirectTwo.(*string))) &&
					assert.Equal(t, unescapedInput, *indirectOne) &&
					assert.Equal(t, unescapedInput, target)

			if !success {
				trace.log(t)
			}
		})
	})
}
