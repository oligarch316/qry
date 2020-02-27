package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

// ===== Error
func TestError(t *testing.T) {
	t.Run("query", queryErrorTests)
	t.Run("field", fieldErrorTests)
	t.Run("key", keyErrorTests)
	t.Run("value list", valueListErrorTests)
	t.Run("value", valueErrorTests)
}

func runCommonErrorTests(t *testing.T, suite decodeErrorSuite) {
	t.Run("root", suite.runRootTests)
	t.Run("literal", suite.runLiteralTests)
	t.Run("faux literal", suite.runFauxLiteralTests)
	t.Run("unmarshaler", suite.runUnmarshalerTests)
}

func queryErrorTests(t *testing.T) {
	suite := newDecodeErrorSuite(qry.LevelQuery)

	if !testing.Short() {
		runCommonErrorTests(t, suite)
		t.Run("unsupported", suite.runUnsupportedCommonTests)
	}

	t.Run("container", func(t *testing.T) {
		if !testing.Short() {
			t.Run("list", func(t *testing.T) { suite.runListTests(t, "&") })
		}

		t.Run("struct", suite.runStructQueryTests)
	})
}

func fieldErrorTests(t *testing.T) {
	suite := newDecodeErrorSuite(qry.LevelField)

	if !testing.Short() {
		runCommonErrorTests(t, suite)
	}

	t.Run("unsupported", func(t *testing.T) {
		if !testing.Short() {
			suite.runUnsupportedCommonTests(t)
		}
		suite.runUnsupportedListTests(t)
	})

	t.Run("container", func(t *testing.T) {
		t.Run("struct", suite.runStructFieldTests)
	})
}

func keyErrorTests(t *testing.T) {
	suite := newDecodeErrorSuite(qry.LevelKey)

	if !testing.Short() {
		runCommonErrorTests(t, suite)
	}

	t.Run("unsupported", func(t *testing.T) {
		if !testing.Short() {
			suite.runUnsupportedCommonTests(t)
		}

		suite.runUnsupportedListTests(t)
		suite.runUnsupportedKeyValTests(t)
	})
}

func valueListErrorTests(t *testing.T) {
	suite := newDecodeErrorSuite(qry.LevelValueList)

	if !testing.Short() {
		runCommonErrorTests(t, suite)
	}

	t.Run("unsupported", func(t *testing.T) {
		if !testing.Short() {
			suite.runUnsupportedCommonTests(t)
		}
		suite.runUnsupportedKeyValTests(t)
	})

	t.Run("container", func(t *testing.T) {
		t.Run("list", func(t *testing.T) { suite.runListTests(t, ",") })
	})
}

func valueErrorTests(t *testing.T) {
	suite := newDecodeErrorSuite(qry.LevelValue)

	runCommonErrorTests(t, suite)
	t.Run("unsupported", func(t *testing.T) {
		suite.runUnsupportedCommonTests(t)
		suite.runUnsupportedListTests(t)
		suite.runUnsupportedKeyValTests(t)
	})
}

// ===== Success
func TestSuccess(t *testing.T) {
	t.Run("query", runQuerySuccessTests)
	t.Run("field", runFieldSuccessTests)
	t.Run("key", runKeySuccessTests)
	t.Run("value list", runValueListSuccessTests)
	t.Run("value", runValueSuccessTests)
}

func runCommonSuccessTests(t *testing.T, suite decodeSuccessSuite) {
	t.Run("literal", suite.runLiteralTests)
	t.Run("faux literal", suite.runFauxLiteralTests)
	t.Run("unmarshaler", suite.runUnmarshalerTests)
}

func runQuerySuccessTests(t *testing.T) {
	suite := newDecodeSuccessSuite(qry.LevelQuery)

	if !testing.Short() {
		runCommonSuccessTests(t, suite)
	}

	t.Run("indirect", func(t *testing.T) {
		if !testing.Short() {
			suite.runIndirectCommonTests(t)
		}

		suite.runIndirectDefaultTests(
			t,
			"key%20A=val%20A1,val%20A2&key%20B=val%20B1,val%20B2",
			map[string][]string{
				"key A": []string{"val A1", "val A2"},
				"key B": []string{"val B1", "val B2"},
			},
		)
	})

	t.Run("container", func(t *testing.T) {
		if !testing.Short() {
			t.Run("list", func(t *testing.T) { suite.runListTests(t, "&") })
		}

		t.Run("map", func(t *testing.T) {
			if !testing.Short() {
				t.Run("single", suite.runMapSingleTests)
			}

			t.Run("multi", suite.runMapMultiTests)
		})

		t.Run("struct", suite.runStructQueryTests)
	})
}

func runFieldSuccessTests(t *testing.T) {
	suite := newDecodeSuccessSuite(qry.LevelField)

	if !testing.Short() {
		runCommonSuccessTests(t, suite)
	}

	t.Run("indirect", func(t *testing.T) {
		if !testing.Short() {
			suite.runIndirectCommonTests(t)
		}

		suite.runIndirectDefaultTests(
			t,
			"key%20A=val%20A1,val%20A2",
			struct {
				Key    string
				Values []string
			}{
				Key:    "key A",
				Values: []string{"val A1", "val A2"},
			},
		)
	})

	t.Run("container", func(t *testing.T) {
		t.Run("map", func(t *testing.T) {
			t.Run("single", suite.runMapSingleTests)
		})

		t.Run("struct", suite.runStructFieldTests)
	})
}

func runKeySuccessTests(t *testing.T) {
	suite := newDecodeSuccessSuite(qry.LevelKey)

	if !testing.Short() {
		runCommonSuccessTests(t, suite)
	}

	t.Run("indirect", func(t *testing.T) {
		if !testing.Short() {
			suite.runIndirectCommonTests(t)
		}

		suite.runIndirectDefaultTests(t, "abc%20xyz", "abc xyz")
	})
}

func runValueListSuccessTests(t *testing.T) {
	suite := newDecodeSuccessSuite(qry.LevelValueList)

	if !testing.Short() {
		runCommonSuccessTests(t, suite)
	}

	t.Run("indirect", func(t *testing.T) {
		if !testing.Short() {
			suite.runIndirectCommonTests(t)
		}

		suite.runIndirectDefaultTests(
			t,
			"val%201,val%202",
			[]string{"val 1", "val 2"},
		)
	})

	t.Run("container", func(t *testing.T) {
		t.Run("list", func(t *testing.T) { suite.runListTests(t, ",") })
	})
}

func runValueSuccessTests(t *testing.T) {
	suite := newDecodeSuccessSuite(qry.LevelValue)

	runCommonSuccessTests(t, suite)
	t.Run("indirect", func(t *testing.T) {
		suite.runIndirectCommonTests(t)
		suite.runIndirectDefaultTests(t, "abc%20xyz", "abc xyz")
	})
}
