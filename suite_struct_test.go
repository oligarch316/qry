package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type structSuiteBase struct{ replaceMode, updateMode decodeSuite }

func newStructSuite(level qry.DecodeLevel) interface{ run(*testing.T) } {
	var (
		ds = decodeSuite{
			level:      level,
			decodeOpts: []qry.Option{qry.SetAllLevelsVia(qry.SetAllowLiteral)},
		}

		base = structSuiteBase{
			replaceMode: ds.with(qry.SetLevelVia(level, qry.SetReplaceContainer)),
			updateMode:  ds.with(qry.SetLevelVia(level, qry.SetUpdateContainer)),
		}
	)

	switch level {
	case qry.LevelQuery:
		return structSuiteQuery(base)
	case qry.LevelField:
		return structSuiteField(base)
	}

	return nil
}

// ===== Query

type structSuiteQuery structSuiteBase

func (ssq structSuiteQuery) run(t *testing.T) {
	t.Run("replace", ssq.runReplaceSubtests)
	t.Run("update", ssq.runUpdateSubtests)
	t.Run("omitted", ssq.runOmittedSubtest)
	t.Run("explicit embed", ssq.runExplicitEmbedSubtests)
	t.Run("implicit embed", ssq.runImplicitEmbedSubtests)
	t.Run("unmarshaler", ssq.runUnmarshalerSubtests)
}

// ----- Basics

type (
	tStructBasic struct {
		KeyA, KeyB string
		KeyC, KeyD *string
		KeyZ       string
	}
	tStructBasicTagName struct {
		Key1  string  `qry:"keyA"`
		Key2  string  `qry:"keyB"`
		Key3  *string `qry:"keyC"`
		Key4  *string `qry:"keyD"`
		Key26 string  `qry:"keyZ"`
	}
)

func (tsb tStructBasic) assert(t *testing.T, a, b, c, d, z string) bool {
	res := assert.Equal(t, a, tsb.KeyA)
	res = assert.Equal(t, b, tsb.KeyB) && res
	res = assert.Equal(t, c, *tsb.KeyC) && res
	res = assert.Equal(t, d, *tsb.KeyD) && res
	return assert.Equal(t, z, tsb.KeyZ) && res
}

func (tsbtn tStructBasicTagName) assert(t *testing.T, one, two, three, four, twentySix string) bool {
	res := assert.Equal(t, one, tsbtn.Key1)
	res = assert.Equal(t, two, tsbtn.Key2) && res
	res = assert.Equal(t, three, *tsbtn.Key3) && res
	res = assert.Equal(t, four, *tsbtn.Key4) && res
	return assert.Equal(t, twentySix, tsbtn.Key26) && res
}

func (ssq structSuiteQuery) runReplaceSubtests(t *testing.T) {
	input := "keyA=val%20A&keyB=val%20B&keyC=val%20C&keyD=val%20D&keyExtra=val%20Extra"

	ssq.replaceMode.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalC, originalD = "orig C", "orig D"
			target               tStructBasic
		)

		target.KeyA, target.KeyB, target.KeyZ = "orig A", "orig B", "orig Z"
		target.KeyC, target.KeyD = &originalC, &originalD

		require.NoError(t, decode(input, &target))
		if target.assert(t, "val A", "val B", "val C", "val D", "") {
			assert.Equal(t, "orig C", originalC)
			assert.Equal(t, "orig D", originalD)
		}
	})

	ssq.replaceMode.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			original3, original4 = "orig 3", "orig 4"
			target               tStructBasicTagName
		)

		target.Key1, target.Key2, target.Key26 = "orig 1", "orig 2", "orig 26"
		target.Key3, target.Key4 = &original3, &original4

		require.NoError(t, decode(input, &target))
		if target.assert(t, "val A", "val B", "val C", "val D", "") {
			assert.Equal(t, "orig 3", original3)
			assert.Equal(t, "orig 4", original4)
		}
	})
}

func (ssq structSuiteQuery) runUpdateSubtests(t *testing.T) {
	input := "keyA=val%20A&keyB=val%20B&keyC=val%20C&keyD=val%20D&keyExtra=val%20Extra"

	ssq.updateMode.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalC, originalD = "orig C", "orig D"
			target               tStructBasic
		)

		target.KeyA, target.KeyB, target.KeyZ = "orig A", "orig B", "orig Z"
		target.KeyC, target.KeyD = &originalC, &originalD

		require.NoError(t, decode(input, &target))
		if target.assert(t, "val A", "val B", "val C", "val D", "orig Z") {
			assert.Equal(t, "val C", originalC)
			assert.Equal(t, "val D", originalD)
		}
	})

	ssq.updateMode.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			original3, original4 = "orig 3", "orig 4"
			target               tStructBasicTagName
		)

		target.Key1, target.Key2, target.Key26 = "orig 1", "orig 2", "orig 26"
		target.Key3, target.Key4 = &original3, &original4

		require.NoError(t, decode(input, &target))
		if target.assert(t, "val A", "val B", "val C", "val D", "orig 26") {
			assert.Equal(t, "val C", original3)
			assert.Equal(t, "val D", original4)
		}
	})

	ssq.updateMode.runSubtest(t, "zero basic", func(t *testing.T, decode tDecode) {
		var target tStructBasic
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "val B", "val C", "val D", "")
	})
}

// ----- Omitted

func (ssq structSuiteQuery) runOmittedSubtest(t *testing.T) {
	ssq.replaceMode.runTest(t, func(t *testing.T, decode tDecode) {
		var (
			input  = "keyA=val%20A&keyB=val%20B"
			target struct {
				KeyA string `qry:"-"`
				KeyB string `qry:"-"`
			}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "", target.KeyA)
		assert.Equal(t, "", target.KeyB)
	})
}

// ----- Explicit embed (via struct tag)

type tStructEmbedded struct{ KeyA, KeyB string }

func (tse tStructEmbedded) assert(t *testing.T, expectedA, expectedB string, msgAndArgs ...interface{}) bool {
	res := assert.Equal(t, expectedA, tse.KeyA, msgAndArgs...)
	return assert.Equal(t, expectedB, tse.KeyB, msgAndArgs...) && res
}

func (ssq structSuiteQuery) runExplicitEmbedSubtests(t *testing.T) {
	var (
		input      = "keyA=val%20A"
		runSubtest = ssq.updateMode.runSubtest
	)

	runSubtest(t, "embed anonymous unexported struct", func(t *testing.T, decode tDecode) {
		var target struct {
			tStructEmbedded `qry:",embed"`
		}
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "")
	})

	runSubtest(t, "embed exported struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB string
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "", target.Embedded.KeyB)
	})

	runSubtest(t, "embed exported zero *struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded *struct {
				KeyA, KeyB string
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "", target.Embedded.KeyB)
	})

	runSubtest(t, "embed exported non-zero *struct", func(t *testing.T, decode tDecode) {
		var (
			original = struct{ KeyA, KeyB string }{"orig A", "orig B"}
			target   = struct {
				Embedded *struct {
					KeyA, KeyB string
				} `qry:",embed"`
			}{Embedded: &original}
		)

		require.NoError(t, decode(input, &target))

		targetSuccess := assert.Equal(t, "val A", target.Embedded.KeyA, "check target")
		targetSuccess = assert.Equal(t, "orig B", target.Embedded.KeyB, "check target") && targetSuccess

		if targetSuccess {
			assert.Equal(t, "val A", original.KeyA, "check original")
			assert.Equal(t, "orig B", original.KeyB, "check original")
		}
	})

	runSubtest(t, "embed struct chain", func(t *testing.T, decode tDecode) {
		var target struct {
			EmbeddedOne struct {
				EmbeddedTwo struct {
					KeyA, KeyB string
				} `qry:",embed"`
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.EmbeddedOne.EmbeddedTwo.KeyA)
		assert.Equal(t, "", target.EmbeddedOne.EmbeddedTwo.KeyB)
	})
}

// ----- Implicit embed

type (
	TStructEmbedded       struct{ KeyA, KeyB string }
	TStructAnonymousChain struct{ TStructEmbedded }
)

func (tse TStructEmbedded) assert(t *testing.T, expectedA, expectedB string, msgAndArgs ...interface{}) bool {
	res := assert.Equal(t, expectedA, tse.KeyA, msgAndArgs...)
	return assert.Equal(t, expectedB, tse.KeyB, msgAndArgs...) && res
}

func (ssq structSuiteQuery) runImplicitEmbedSubtests(t *testing.T) {
	var (
		input      = "keyA=val%20A"
		runSubtest = ssq.updateMode.runSubtest
	)

	runSubtest(t, "embed anonymous unexported struct", func(t *testing.T, decode tDecode) {
		var target struct{ tStructEmbedded }
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "")
	})

	runSubtest(t, "embed anonymous exported struct", func(t *testing.T, decode tDecode) {
		var target struct{ TStructEmbedded }
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "")
	})

	runSubtest(t, "embed anonymous exported zero *struct", func(t *testing.T, decode tDecode) {
		var target struct{ *TStructEmbedded }
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "")
	})

	runSubtest(t, "embed anonymous exported non-zero *struct", func(t *testing.T, decode tDecode) {
		var (
			original = TStructEmbedded{KeyA: "orig A", KeyB: "orig B"}
			target   = struct{ *TStructEmbedded }{&original}
		)

		require.NoError(t, decode(input, &target))
		if target.assert(t, "val A", "orig B", "check target") {
			original.assert(t, "val A", "orig B", "check original")
		}
	})

	runSubtest(t, "embed anonymous struct chain", func(t *testing.T, decode tDecode) {
		var target struct{ TStructAnonymousChain }
		require.NoError(t, decode(input, &target))
		target.assert(t, "val A", "")
	})
}

// ----- "Tricky" unmarshaler cases

type (
	tStructEmbeddedUnmarshaler struct{ KeyC string }
	TStructEmbeddedUnmarshaler struct{ KeyC string }
)

func (tseu *tStructEmbeddedUnmarshaler) UnmarshalText(_ []byte) error {
	tseu.KeyC = "unmarshalText called"
	return nil
}

func (tseu *TStructEmbeddedUnmarshaler) UnmarshalText(_ []byte) error {
	tseu.KeyC = "unmarshalText called"
	return nil
}

func (ssq structSuiteQuery) runUnmarshalerSubtests(t *testing.T) {
	var (
		input      = "keyA=val%20A&keyB=val%20B&keyC=val%20C"
		runSubtest = ssq.updateMode.runSubtest
	)

	runSubtest(t, "pathological anonymous exported unmarshaler", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				TStructEmbeddedUnmarshaler `qry:"keyC"`
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "unmarshalText called", target.Embedded.KeyC)
	})

	// TODO: !!!! Turn this into an error test !!!!
	// runSubtest(t, "pathological anonymous unexported unmarshaler", func(t *testing.T, decode tDecode) {
	//     var target struct {
	//         Embedded struct {
	//             KeyA, KeyB string
	//             tStructEmbeddedUnmarshaler `qry:"keyC"`
	//         } `qry:",embed"`
	//     }
	//
	//     require.NoError(t, decode(input, &target))
	//     assert.Equal(t, "val A", target.Embedded.KeyA)
	//     assert.Equal(t, "val B", target.Embedded.KeyB)
	//     assert.Equal(t, "TODO", target.Embedded.KeyC)
	// })

	runSubtest(t, "pathological anonymous exported embedded unmarshaler", func(y *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				TStructEmbeddedUnmarshaler `qry:",embed"`
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "val C", target.Embedded.KeyC)
	})

	runSubtest(t, "pathological anonymous unexported embedded unmarshaler", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				tStructEmbeddedUnmarshaler `qry:",embed"`
			} `qry:",embed"`
		}

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "val C", target.Embedded.KeyC)
	})
}

// ===== Field

type structSuiteField structSuiteBase

func (ssf structSuiteField) run(t *testing.T) {
	// NOTE:
	// Just run replace and update for fields, leave fancier scenarios to the
	// query level tests.
	//
	// This shortcut is no longer OK when/if structParser behavior changes
	// based on DecodeLevel.

	t.Run("replace", ssf.runReplaceSubtests)
	t.Run("update", ssf.runUpdateSubtests)
}

type (
	tStructFieldBasic        struct{ Key, Values *string }
	tStructFieldBasicTagName struct {
		MyKey    *string `qry:"key"`
		MyValues *string `qry:"values"`
	}
)

func (tsfb tStructFieldBasic) assert(t *testing.T, expectedKey, expectedVals string) bool {
	if !assert.NotNil(t, tsfb.Key) || !assert.NotNil(t, tsfb.Values) {
		return false
	}

	res := assert.Equal(t, expectedKey, *tsfb.Key)
	return assert.Equal(t, expectedVals, *tsfb.Values) && res
}

func (tsfbtn tStructFieldBasicTagName) assert(t *testing.T, expectedKey, expectedVals string) bool {
	if !assert.NotNil(t, tsfbtn.MyKey) || !assert.NotNil(t, tsfbtn.MyValues) {
		return false
	}

	res := assert.Equal(t, expectedKey, *tsfbtn.MyKey)
	return assert.Equal(t, expectedVals, *tsfbtn.MyValues) && res
}

func (ssf structSuiteField) runReplaceSubtests(t *testing.T) {
	input := "key%20A=val%20A"

	ssf.replaceMode.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasic{Key: &originalKey, Values: &originalVals}
		)

		require.NoError(t, decode(input, &target))
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "orig key", originalKey)
			assert.Equal(t, "orig vals", originalVals)
		}
	})

	ssf.replaceMode.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasicTagName{MyKey: &originalKey, MyValues: &originalVals}
		)

		require.NoError(t, decode(input, &target))
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "orig key", originalKey)
			assert.Equal(t, "orig vals", originalVals)
		}
	})
}

func (ssf structSuiteField) runUpdateSubtests(t *testing.T) {
	input := "key%20A=val%20A"

	ssf.updateMode.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasic{Key: &originalKey, Values: &originalVals}
		)

		require.NoError(t, decode(input, &target))
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "key A", originalKey)
			assert.Equal(t, "val A", originalVals)
		}
	})

	ssf.updateMode.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasicTagName{MyKey: &originalKey, MyValues: &originalVals}
		)

		require.NoError(t, decode(input, &target))
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "key A", originalKey)
			assert.Equal(t, "val A", originalVals)
		}
	})

	ssf.updateMode.runSubtest(t, "zero basic", func(t *testing.T, decode tDecode) {
		var target tStructFieldBasic
		require.NoError(t, decode(input, &target))
		target.assert(t, "key A", "val A")
	})
}
