package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

// ===== New hotness
func commonSubtests(t *testing.T, level qry.DecodeLevel, skipOnShort bool) {
	if skipOnShort && testing.Short() {
		return
	}

	t.Run("literal", newLiteralSuite(level).run)
	t.Run("faux literal", newFauxLiteralSuite(level).run)
	t.Run("unmarshaler", newUnmarshalerSuite(level).run)
}

func querySubtests(t *testing.T) {
	var (
		indirectSuite = newIndirectSuite(
			qry.LevelQuery,
			"key%20A=val%20A1,val%20A2&key%20B=val%20B1,val%20B2",
			map[string][]string{
				"key A": []string{"val A1", "val A2"},
				"key B": []string{"val B1", "val B2"},
			},
			true,
		)
		listSuite   = newListSuite(qry.LevelQuery, "&")
		mapSuite    = newMapSuite(qry.LevelQuery)
		structSuite = newStructSuite(qry.LevelQuery)
	)

	commonSubtests(t, qry.LevelQuery, true)
	t.Run("indirect", indirectSuite.run)
	t.Run("container", func(t *testing.T) {
		if !testing.Short() {
			t.Run("list", listSuite.run)
		}

		t.Run("map", func(t *testing.T) {
			if !testing.Short() {
				mapSuite.run(t)
			}
			mapSuite.runMulti(t)
		})

		t.Run("struct", structSuite.run)
	})
}

func fieldSubtests(t *testing.T) {
	var (
		indirectSuite = newIndirectSuite(
			qry.LevelField,
			"key%20A=val%20A1,val%20A2",
			struct {
				Key    string
				Values []string
			}{
				Key:    "key A",
				Values: []string{"val A1", "val A2"},
			},
			true,
		)
		mapSuite    = newMapSuite(qry.LevelField)
		structSuite = newStructSuite(qry.LevelField)
	)

	commonSubtests(t, qry.LevelField, true)
	t.Run("indirect", indirectSuite.run)
	t.Run("container", func(t *testing.T) {
		t.Run("map", mapSuite.run)
		t.Run("struct", structSuite.run)
	})
}

func keySubtests(t *testing.T) {
	indirectSuite := newIndirectSuite(
		qry.LevelKey,
		"abc%20xyz",
		"abc xyz",
		true,
	)

	commonSubtests(t, qry.LevelKey, true)
	t.Run("indirect", indirectSuite.run)
}

func valueListSubtests(t *testing.T) {
	var (
		indirectSuite = newIndirectSuite(
			qry.LevelValueList,
			"val%201,val%202",
			[]string{"val 1", "val 2"},
			true,
		)
		listSuite = newListSuite(qry.LevelValueList, ",")
	)

	commonSubtests(t, qry.LevelValueList, true)
	t.Run("indirect", indirectSuite.run)
	t.Run("container", func(t *testing.T) { t.Run("list", listSuite.run) })
}

func valueSubtests(t *testing.T) {
	indirectSuite := newIndirectSuite(
		qry.LevelValue,
		"abc%20xyz",
		"abc xyz",
		false,
	)

	commonSubtests(t, qry.LevelValue, false)
	t.Run("indirect", indirectSuite.run)
}

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
