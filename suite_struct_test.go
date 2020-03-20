package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

// ===== Error
type (
	tStructError            struct{ Key string }
	tStructErrorUnmarshaler struct{}
)

func (tsqeu *tStructErrorUnmarshaler) UnmarshalText(_ []byte) error { return nil }

func (des decodeErrorSuite) runStructUnescapeTests(t *testing.T) {
	des.withUnescapeError("forced unescape error").runSubtest(t, "unescape error", func(t *testing.T, decode tDecode) {
		var target tStructError
		actual := decode("xyz", &target)
		assertErrorMessage(t, "forced unescape error", actual)
	})
}

func (des decodeErrorSuite) runStructParseTests(t *testing.T) {
	t.Run("tag", des.runStructParseTagSubtests)
	t.Run("explicit embed", des.runStructParseExplicitEmbedSubtests)
	t.Run("tagged invalid", des.runStructParseTaggedInvalidSubtests)
}

func (des decodeErrorSuite) runStructParseTagSubtests(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		des.runSubtest(t, "empty", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qry:""`
			}
			actual := decode("xyz", &target)
			assertErrorMessage(t, "empty base tag", actual)
		})

		des.runSubtest(t, "unknown directive", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qry:",nonDirective"`
			}
			actual := decode("xyz", &target)
			assertErrorMessage(t, "invalid base tag directive 'nonDirective'", actual)
		})

		des.runSubtest(t, "embed and non-empty name", func(t *testing.T, decode tDecode) {
			var target struct {
				Embedded struct{ Key string } `qry:"keyA,embed"`
			}
			actual := decode("xyz", &target)
			assertErrorMessage(t, "mutually exclusive base tag directive 'embed' and non-empty name", actual)
		})
	})

	t.Run("set", func(t *testing.T) {
		des.runSubtest(t, "empty", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qrySet:""`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "empty set tag", actual)
		})

		des.runSubtest(t, "unknown default option", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qrySet:"nonSetOpt"`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "invalid set tag option 'nonSetOpt'", actual)
		})

		des.runSubtest(t, "unknown explicit option", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qrySet:"valueList=nonSetOpt"`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "invalid set tag option 'nonSetOpt'", actual)
		})

		des.runSubtest(t, "unknown explicit level", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qrySet:"nonLevel=allowLiteral"`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "invalid set tag level 'nonLevel'", actual)
		})
	})

	t.Run("combined", func(t *testing.T) {
		des.runSubtest(t, "omit and set options", func(t *testing.T, decode tDecode) {
			var target struct {
				Key string `qry:"-" qrySet:"allowLiteral"`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "mutually exclusive base tag name '-' (omit) and set tag options", actual)
		})

		des.runSubtest(t, "embed and set options", func(t *testing.T, decode tDecode) {
			var target struct {
				Embedded struct{ Key string } `qry:",embed" qrySet:"allowLiteral"`
			}
			actual := decode("xzy", &target)
			assertErrorMessage(t, "mutually exclusive base tag directive 'embed' and set tag options", actual)
		})
	})
}

func (des decodeErrorSuite) runStructParseExplicitEmbedSubtests(t *testing.T) {
	des.runSubtest(t, "non-anonymous unexported", func(t *testing.T, decode tDecode) {
		var target struct {
			embedded struct{ Key string } `qry:",embed"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, "'embed' directive on non-anonymous unexported field", actual)
	})

	des.runSubtest(t, "anonymous unexported pointer", func(t *testing.T, decode tDecode) {
		var target struct {
			*tStructError `qry:",embed"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, "'embed' directive on unexported pointer field", actual)
	})

	des.runSubtest(t, "non-pointer non-struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Key string `qry:",embed"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, "'embed' directive on invalid type", actual)
	})

	des.runSubtest(t, "pointer to non-struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded *string `qry:",embed"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, "'embed' directive on invalid type", actual)
	})
}

func (des decodeErrorSuite) runStructParseTaggedInvalidSubtests(t *testing.T) {
	expected := "tag on unexported field"

	des.runSubtest(t, "non-anonymous unexported", func(t *testing.T, decode tDecode) {
		var target struct {
			myKey string `qry:"key"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "anonymous unexported pointer", func(t *testing.T, decode tDecode) {
		var target struct {
			*tStructError `qry:"key"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "pathological anonymous unexported unmarshaler", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				tStructErrorUnmarshaler `qry:"key"`
			} `qry:",embed"`
		}
		actual := decode("xyz", &target)
		assertErrorMessage(t, expected, actual)
	})
}

// ===== Success

// ----- Query
func (dss decodeSuccessSuite) runStructQueryTests(t *testing.T) {
	t.Run("replace", dss.runStructQueryReplaceSubtests)
	t.Run("update", dss.runStructQueryUpdateSubtests)
	t.Run("omitted", dss.runStructQueryOmittedSubtest)
	t.Run("explicit embed", dss.runStructQueryExplicitEmbedSubtests)
	t.Run("implicit embed", dss.runStructQueryImplicitEmbedSubtests)
	t.Run("pathological", dss.runStructQueryPathologicalSubtests)
}

// > Basics
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

func (dss decodeSuccessSuite) runStructQueryReplaceSubtests(t *testing.T) {
	var (
		input  = "keyA=val%20A&keyB=val%20B&keyC=val%20C&keyD=val%20D&keyExtra=val%20Extra"
		runner = dss.withSetOpts(qry.SetReplaceContainer)
	)

	runner.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalC, originalD = "orig C", "orig D"
			target               tStructBasic
		)

		target.KeyA, target.KeyB, target.KeyZ = "orig A", "orig B", "orig Z"
		target.KeyC, target.KeyD = &originalC, &originalD

		decode(input, &target)
		if target.assert(t, "val A", "val B", "val C", "val D", "") {
			assert.Equal(t, "orig C", originalC)
			assert.Equal(t, "orig D", originalD)
		}
	})

	runner.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			original3, original4 = "orig 3", "orig 4"
			target               tStructBasicTagName
		)

		target.Key1, target.Key2, target.Key26 = "orig 1", "orig 2", "orig 26"
		target.Key3, target.Key4 = &original3, &original4

		decode(input, &target)
		if target.assert(t, "val A", "val B", "val C", "val D", "") {
			assert.Equal(t, "orig 3", original3)
			assert.Equal(t, "orig 4", original4)
		}
	})
}

func (dss decodeSuccessSuite) runStructQueryUpdateSubtests(t *testing.T) {
	var (
		input  = "keyA=val%20A&keyB=val%20B&keyC=val%20C&keyD=val%20D&keyExtra=val%20Extra"
		runner = dss.withSetOpts(qry.SetUpdateContainer)
	)

	runner.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalC, originalD = "orig C", "orig D"
			target               tStructBasic
		)

		target.KeyA, target.KeyB, target.KeyZ = "orig A", "orig B", "orig Z"
		target.KeyC, target.KeyD = &originalC, &originalD

		decode(input, &target)
		if target.assert(t, "val A", "val B", "val C", "val D", "orig Z") {
			assert.Equal(t, "val C", originalC)
			assert.Equal(t, "val D", originalD)
		}
	})

	runner.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			original3, original4 = "orig 3", "orig 4"
			target               tStructBasicTagName
		)

		target.Key1, target.Key2, target.Key26 = "orig 1", "orig 2", "orig 26"
		target.Key3, target.Key4 = &original3, &original4

		decode(input, &target)
		if target.assert(t, "val A", "val B", "val C", "val D", "orig 26") {
			assert.Equal(t, "val C", original3)
			assert.Equal(t, "val D", original4)
		}
	})

	runner.runSubtest(t, "zero basic", func(t *testing.T, decode tDecode) {
		var target tStructBasic
		decode(input, &target)
		target.assert(t, "val A", "val B", "val C", "val D", "")
	})
}

// > Omitted
func (dss decodeSuccessSuite) runStructQueryOmittedSubtest(t *testing.T) {
	dss.runTest(t, func(t *testing.T, decode tDecode) {
		var (
			input  = "keyA=val%20A&-=val%20hyphen"
			target struct {
				KeyA string `qry:"-"`
				KeyB string `qry:"-,"`
			}
		)

		decode(input, &target)
		assert.Equal(t, "", target.KeyA)
		assert.Equal(t, "val hyphen", target.KeyB)
	})
}

// > Explicit embed (via struct tag)
type tStructEmbedded struct{ KeyA, KeyB string }

func (tse tStructEmbedded) assert(t *testing.T, expectedA, expectedB string, msgAndArgs ...interface{}) bool {
	res := assert.Equal(t, expectedA, tse.KeyA, msgAndArgs...)
	return assert.Equal(t, expectedB, tse.KeyB, msgAndArgs...) && res
}

func (dss decodeSuccessSuite) runStructQueryExplicitEmbedSubtests(t *testing.T) {
	input := "keyA=val%20A"

	dss.runSubtest(t, "embed anonymous unexported struct", func(t *testing.T, decode tDecode) {
		var target struct {
			tStructEmbedded `qry:",embed"`
		}
		decode(input, &target)
		target.assert(t, "val A", "")
	})

	dss.runSubtest(t, "embed exported struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB string
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "", target.Embedded.KeyB)
	})

	dss.runSubtest(t, "embed exported zero *struct", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded *struct {
				KeyA, KeyB string
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "", target.Embedded.KeyB)
	})

	dss.runSubtest(t, "embed exported non-zero *struct", func(t *testing.T, decode tDecode) {
		var (
			original = struct{ KeyA, KeyB string }{"orig A", "orig B"}
			target   = struct {
				Embedded *struct {
					KeyA, KeyB string
				} `qry:",embed"`
			}{Embedded: &original}
		)

		decode(input, &target)

		targetSuccess := assert.Equal(t, "val A", target.Embedded.KeyA, "check target")
		targetSuccess = assert.Equal(t, "orig B", target.Embedded.KeyB, "check target") && targetSuccess

		if targetSuccess {
			assert.Equal(t, "val A", original.KeyA, "check original")
			assert.Equal(t, "orig B", original.KeyB, "check original")
		}
	})

	dss.runSubtest(t, "embed struct chain", func(t *testing.T, decode tDecode) {
		var target struct {
			EmbeddedOne struct {
				EmbeddedTwo struct {
					KeyA, KeyB string
				} `qry:",embed"`
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.EmbeddedOne.EmbeddedTwo.KeyA)
		assert.Equal(t, "", target.EmbeddedOne.EmbeddedTwo.KeyB)
	})
}

// > Implicit embed
type (
	TStructEmbedded       struct{ KeyA, KeyB string }
	TStructAnonymousChain struct{ TStructEmbedded }
)

func (tse TStructEmbedded) assert(t *testing.T, expectedA, expectedB string, msgAndArgs ...interface{}) bool {
	res := assert.Equal(t, expectedA, tse.KeyA, msgAndArgs...)
	return assert.Equal(t, expectedB, tse.KeyB, msgAndArgs...) && res
}

func (dss decodeSuccessSuite) runStructQueryImplicitEmbedSubtests(t *testing.T) {
	input := "keyA=val%20A"

	dss.runSubtest(t, "embed anonymous unexported struct", func(t *testing.T, decode tDecode) {
		var target struct{ tStructEmbedded }
		decode(input, &target)
		target.assert(t, "val A", "")
	})

	dss.runSubtest(t, "embed anonymous exported struct", func(t *testing.T, decode tDecode) {
		var target struct{ TStructEmbedded }
		decode(input, &target)
		target.assert(t, "val A", "")
	})

	dss.runSubtest(t, "embed anonymous exported zero *struct", func(t *testing.T, decode tDecode) {
		var target struct{ *TStructEmbedded }
		decode(input, &target)
		target.assert(t, "val A", "")
	})

	dss.runSubtest(t, "embed anonymous exported non-zero *struct", func(t *testing.T, decode tDecode) {
		var (
			original = TStructEmbedded{KeyA: "orig A", KeyB: "orig B"}
			target   = struct{ *TStructEmbedded }{&original}
		)

		decode(input, &target)
		if target.assert(t, "val A", "orig B", "check target") {
			original.assert(t, "val A", "orig B", "check original")
		}
	})

	dss.runSubtest(t, "embed anonymous struct chain", func(t *testing.T, decode tDecode) {
		var target struct{ TStructAnonymousChain }
		decode(input, &target)
		target.assert(t, "val A", "")
	})
}

// > Pathological unmarshaler scenarios
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

func (dss decodeSuccessSuite) runStructQueryPathologicalSubtests(t *testing.T) {
	input := "keyA=val%20A&keyB=val%20B&keyC=val%20C"

	dss.runSubtest(t, "anonymous exported unmarshaler", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				TStructEmbeddedUnmarshaler `qry:"keyC"`
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "unmarshalText called", target.Embedded.KeyC)
	})

	dss.runSubtest(t, "anonymous exported embedded unmarshaler", func(y *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				TStructEmbeddedUnmarshaler `qry:",embed"`
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "val C", target.Embedded.KeyC)
	})

	dss.runSubtest(t, "anonymous unexported embedded unmarshaler", func(t *testing.T, decode tDecode) {
		var target struct {
			Embedded struct {
				KeyA, KeyB                 string
				tStructEmbeddedUnmarshaler `qry:",embed"`
			} `qry:",embed"`
		}

		decode(input, &target)
		assert.Equal(t, "val A", target.Embedded.KeyA)
		assert.Equal(t, "val B", target.Embedded.KeyB)
		assert.Equal(t, "val C", target.Embedded.KeyC)
	})
}

// ----- Field
func (dss decodeSuccessSuite) runStructFieldTests(t *testing.T) {
	// NOTE:
	// Just run basics (replace and update) for fields, leave fancier scenarios
	// to the query level tests.
	//
	// This shortcut is no longer OK when/if structParser behavior changes
	// based on DecodeLevel.

	t.Run("replace", dss.runStructFieldReplaceSubtests)
	t.Run("update", dss.runStructFieldUpdateSubtests)
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

func (dss decodeSuccessSuite) runStructFieldReplaceSubtests(t *testing.T) {
	var (
		input  = "key%20A=val%20A"
		runner = dss.withSetOpts(qry.SetReplaceContainer)
	)

	runner.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasic{Key: &originalKey, Values: &originalVals}
		)

		decode(input, &target)
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "orig key", originalKey)
			assert.Equal(t, "orig vals", originalVals)
		}
	})

	runner.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasicTagName{MyKey: &originalKey, MyValues: &originalVals}
		)

		decode(input, &target)
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "orig key", originalKey)
			assert.Equal(t, "orig vals", originalVals)
		}
	})
}

func (dss decodeSuccessSuite) runStructFieldUpdateSubtests(t *testing.T) {
	var (
		input  = "key%20A=val%20A"
		runner = dss.withSetOpts(qry.SetUpdateContainer)
	)

	runner.runSubtest(t, "basic", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasic{Key: &originalKey, Values: &originalVals}
		)

		decode(input, &target)
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "key A", originalKey)
			assert.Equal(t, "val A", originalVals)
		}
	})

	runner.runSubtest(t, "basic tag name", func(t *testing.T, decode tDecode) {
		var (
			originalKey, originalVals = "orig key", "orig vals"
			target                    = tStructFieldBasicTagName{MyKey: &originalKey, MyValues: &originalVals}
		)

		decode(input, &target)
		if target.assert(t, "key A", "val A") {
			assert.Equal(t, "key A", originalKey)
			assert.Equal(t, "val A", originalVals)
		}
	})

	runner.runSubtest(t, "zero basic", func(t *testing.T, decode tDecode) {
		var target tStructFieldBasic
		decode(input, &target)
		target.assert(t, "key A", "val A")
	})
}
