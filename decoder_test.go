package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

// ===== New hotness
func commonSubtests(t *testing.T, level qry.DecodeLevel, skipOnShort bool) {
	if skipOnShort && testing.Short() {
		return
	}

	t.Run("literal", newLiteralSuite(level).subtests)
	t.Run("faux literal", newFauxLiteralSuite(level).subtests)
	t.Run("unmarshaler", func(t *testing.T) { t.Skip("TODO") })
	t.Run("indirect", func(t *testing.T) { t.Skip("TODO") })
}

func querySubtests(t *testing.T) {
	commonSubtests(t, qry.LevelQuery, true)
	t.Run("container", func(t *testing.T) { t.Skip("TODO") })
}

func fieldSubtests(t *testing.T) {
	commonSubtests(t, qry.LevelField, true)
	t.Run("container", func(t *testing.T) { t.Skip("TODO") })
}

func keySubtests(t *testing.T) { commonSubtests(t, qry.LevelKey, true) }

func valueListSubtests(t *testing.T) {
	commonSubtests(t, qry.LevelValueList, true)
	t.Run("container", func(t *testing.T) { t.Skip("TODO") })
}

func valueSubtests(t *testing.T) { commonSubtests(t, qry.LevelValue, false) }

func TestError(t *testing.T) {
	t.Skip("TODO")
}

func TestSuccess(t *testing.T) {
	t.Run("query", querySubtests)
	t.Run("field", fieldSubtests)
	t.Run("key", keySubtests)
	t.Run("value list", valueListSubtests)
	t.Run("value", valueSubtests)
}

// ===== Old and busted

func TestQueryError(t *testing.T) {
	t.Run("root", func(t *testing.T) { testRootErrors(t, qry.LevelQuery) })
	t.Run("literal", func(t *testing.T) { testLiteralErrors(t, qry.LevelQuery) })
	t.Run("faux literal", func(t *testing.T) { testFauxLiteralErrors(t, qry.LevelQuery) })
	t.Run("container", func(t *testing.T) {
		t.Run("array", func(t *testing.T) { testArrayErrors(t, qry.LevelQuery) })
		t.Run("struct", func(t *testing.T) {
			testStructTagErrors(t, qry.LevelQuery)
			t.Run("key unescape", func(t *testing.T) {
				t.Skip("TODO: converter.Unescape error")
			})
		})
	})
	t.Run("unsupported", func(t *testing.T) {
		testCommonUnsupportedErrors(t, qry.LevelQuery)

		// TODO: All literals when set mode is "disallow"
	})
}

func TestFieldError(t *testing.T) {
	t.Skip("TODO")
}

func TestKeyError(t *testing.T) {
	t.Skip("TODO")
}

func TestValueListError(t *testing.T) {
	t.Skip("TODO")
}

func TestValueError(t *testing.T) {
	t.Skip("TODO")
}

func TestQuerySuccess(t *testing.T) {
	t.Run("literal", func(t *testing.T) { testLiterals(t, qry.LevelQuery) })
	t.Run("faux literal", func(t *testing.T) { testFauxLiterals(t, qry.LevelQuery) })
	t.Run("indirect", func(t *testing.T) { testIndirects(t, qry.LevelQuery) })
	t.Run("container", func(t *testing.T) {
		// TODO!!!: Test update mode for containers

		base := newTest(
			configOptionsAs(qry.SetAllLevelsVia(qry.SetAllowLiteral)),
			decodeLevelAs(qry.LevelQuery),
		)

		t.Run("[]string target", func(t *testing.T) {
			var (
				input    = "field%20A&field%20B"
				expected = []string{"field A", "field B"}
				target   []string
			)

			trace := base.with(inputAs(input)).require(t, &target)
			if !assert.Equal(t, expected, target) {
				trace.log(t)
			}
		})

		t.Run("[2]string target", func(t *testing.T) {
			var (
				input    = "field%20A&field%20B"
				expected = [2]string{"field A", "field B"}
				target   [2]string
			)

			trace := base.with(inputAs(input)).require(t, &target)
			if !assert.Equal(t, expected, target) {
				trace.log(t)
			}
		})

		t.Run("map[string]string target", func(t *testing.T) {
			var (
				input    = "key%20A=vals%20A&key%20B=vals%20B"
				expected = map[string]string{
					"key A": "vals A",
					"key B": "vals B",
				}
				target map[string]string
			)

			trace := base.with(inputAs(input)).require(t, &target)
			if !assert.Equal(t, expected, target) {
				trace.log(t)
			}
		})

		t.Run("struct", testStructs)

		// TODO: testing key unescaping during struct decoding
	})
}

func TestFieldSuccess(t *testing.T) {
	t.Skip("TODO")
}

func TestKeySuccess(t *testing.T) {
	t.Skip("TODO")
}

func TestValueListSuccess(t *testing.T) {
	t.Skip("TODO")
}

func TestValueSuccess(t *testing.T) {
	t.Skip("TODO")
}

func testRootErrors(t *testing.T, level qry.DecodeLevel) {
	base := newTest(
		configOptionsAs(qry.SetLevelVia(level, qry.SetAllowLiteral)),
		decodeLevelAs(level),
		inputAs("xyz"),
		checkDecodeError(
			assertDecodeLevel(qry.LevelRoot),
		),
	)

	t.Run("non-pointer target", func(t *testing.T) {
		var target string
		base.with(errorMessageAs("non-pointer target")).require(t, target)
	})

	t.Run("nil pointer target", func(t *testing.T) {
		var target *string
		base.with(errorMessageAs("nil pointer target")).require(t, target)
	})
}

func testCommonUnsupportedErrors(t *testing.T, level qry.DecodeLevel) {
	unsupportedTest := newTest(
		configOptionsAs(qry.SetLevelVia(level, qry.SetAllowLiteral)),
		decodeLevelAs(level),
		inputAs("xyz"),
		checkDecodeError(
			assertDecodeLevel(level),
		),
		errorMessageAs("unsupported target type"),
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

func testArrayErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("too small", func(t *testing.T) {
		t.Skip("TODO: insufficient target length error")
	})
}
