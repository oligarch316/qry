package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

func TestQueryError(t *testing.T) {
	t.Run("root", func(t *testing.T) { testRootErrors(t, qry.LevelQuery) })
	t.Run("literals", func(t *testing.T) { testLiteralErrors(t, qry.LevelQuery) })
	t.Run("unsupported", func(t *testing.T) { testCommonUnsupportedErrors(t, qry.LevelQuery) })
}

// TODO: Field/Key/ValueList/Value Error

func TestQuerySuccess(t *testing.T) {
	t.Run("literal", func(t *testing.T) {
		testLiterals(t, qry.LevelQuery)
		testFauxLiterals(t, qry.LevelQuery)
	})
	t.Run("indirect", func(t *testing.T) { testIndirects(t, qry.LevelQuery) })
	t.Run("container", func(t *testing.T) {
		t.Skip("TODO")
		// TODO
	})
}

// TODO: Field/Key/ValueList/Value Success

func testRootErrors(t *testing.T, level qry.DecodeLevel) {
	base := newTest(
		configOptionsAs(qry.SetOptionsAs(level, qry.SetAllowLiteral)),
		decodeLevelAs(level),
		inputAs("xyz"),
		errorLevelAs(qry.LevelRoot),
	)

	t.Run("non-pointer target", func(t *testing.T) {
		var target string
		base.with(errorAs("non-pointer target")).require(t, target)
	})

	t.Run("nil pointer target", func(t *testing.T) {
		var target *string
		base.with(errorAs("nil pointer target")).require(t, target)
	})
}

func testCommonUnsupportedErrors(t *testing.T, level qry.DecodeLevel) {
	unsupportedTest := newTest(
		configOptionsAs(qry.SetOptionsAs(level, qry.SetAllowLiteral)),
		decodeLevelAs(level),
		inputAs("xyz"),
		errorLevelAs(level),
		errorAs("unsupported target type"),
	)

	t.Run("chan target", func(t *testing.T) {
		var target chan struct{}
		unsupportedTest.require(t, &target)
	})

	t.Run("func target", func(t *testing.T) {
		var target func()
		unsupportedTest.require(t, &target)
	})
}

func testLiteralErrors(t *testing.T, level qry.DecodeLevel) {
	t.Skip("TODO")

	// converter.Escape()
	// all possible strconv.XYZ errors
	// unmarshaler errors
}

func testLiterals(t *testing.T, level qry.DecodeLevel) {
	base := newTest(
		configOptionsAs(qry.SetOptionsAs(level, qry.SetAllowLiteral)),
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

func testIndirects(t *testing.T, level qry.DecodeLevel) {
	var (
		rawInput       = "abc%20xyz"
		unescapedInput = "abc xyz"

		base = newTest(
			configOptionsAs(qry.SetOptionsAs(level, qry.SetAllowLiteral)),
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
