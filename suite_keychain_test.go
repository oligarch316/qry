package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Error

// NOTE:
// Unescape and structparser error testing when decoding structs is left
// to the "struct family" of error tests, despite occuring within the
// decodeKeyChain(...) function.

func (des decodeErrorSuite) runKeyChainTests(t *testing.T) {
	var (
		input  = "keyA.keyX=val%20AX"
		runner = des.with(
			qry.SeparateKeyChainBy('.'),
			qry.IgnoreInvalidKeys(false),
		)
	)

	runner.runSubtest(t, "non-indexable target error", func(t *testing.T, decode tDecode) {
		var target map[string]string
		actual := decode(input, &target)
		assertErrorMessage(t, "non-indexable key chain target", actual)
	})

	runner.runSubtest(t, "unknown key error", func(t *testing.T, decode tDecode) {
		var target map[string]struct{ KeyOther string }
		actual := decode(input, &target)
		assertErrorMessage(t, "unknown key", actual)
	})
}

// ===== Success

// ----- Query
func (dss decodeSuccessSuite) runKeyChainQueryTests(t *testing.T) {
	t.Run("map chain", dss.runKeyChainQueryMapSubtests)
	t.Run("struct chain", dss.runKeyChainQueryStructSubtests)
	t.Run("mixed chain", dss.runKeyChainQueryMixedSubtests)
}

func (dss decodeSuccessSuite) runKeyChainQueryMapSubtests(t *testing.T) {
	var (
		input  = "keyA.keyX=val%20AX&keyB.keyX=val%20BX"
		runner = dss.withKeyChainSep('.')
	)

	runner.runSubtest(t, "map[string]map[string]string", func(t *testing.T, decode tDecode) {
		var (
			target = map[string]map[string]string{
				"keyA": map[string]string{"keyX": "orig AX"},
				"keyB": map[string]string{"keyX": "orig BX"},
			}
			expected = map[string]map[string]string{
				"keyA": map[string]string{"keyX": "val AX"},
				"keyB": map[string]string{"keyX": "val BX"},
			}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})

	runner.runSubtest(t, "map[string]map[string]*string", func(t *testing.T, decode tDecode) {
		var (
			originalBX = "orig BX"
			originalBY = "orig BY"
			target     = map[string]map[string]*string{
				"keyB": map[string]*string{
					"keyX": &originalBX,
					"keyY": &originalBY,
				},
			}

			expectedAX = "val AX"
			expectedBX = "val BX"
			expectedBY = "orig BY"
			expected   = map[string]map[string]*string{
				"keyA": map[string]*string{"keyX": &expectedAX},
				"keyB": map[string]*string{
					"keyX": &expectedBX,
					"keyY": &expectedBY,
				},
			}
		)

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "val BX", originalBX)
		assert.Equal(t, "orig BY", originalBY)
	})

	runner.runSubtest(t, "map[string]*map[string]*string", func(t *testing.T, decode tDecode) {
		var (
			originalBX, originalBY = "orig BX", "orig BY"
			originalB              = map[string]*string{
				"keyX": &originalBX,
				"keyY": &originalBY,
			}
			target = map[string]*map[string]*string{"keyB": &originalB}

			expectedAX             = "val AX"
			expectedA              = map[string]*string{"keyX": &expectedAX}
			expectedBX, expectedBY = "val BX", "orig BY"
			expectedB              = map[string]*string{
				"keyX": &expectedBX,
				"keyY": &expectedBY,
			}
			expected = map[string]*map[string]*string{
				"keyA": &expectedA,
				"keyB": &expectedB,
			}
		)

		decode(input, &target)
		require.Equal(t, expected, target)
		require.Equal(t, expectedB, originalB)
		assert.Equal(t, "val BX", originalBX)
		assert.Equal(t, "orig BY", originalBY)
	})
}

type (
	tKeyChainQueryXY struct {
		KeyX string
		KeyY *string
	}

	tKeyChainQueryEmbeddedC struct{ KeyC tKeyChainQueryXY }

	tKeyChainQueryAB struct {
		KeyA tKeyChainQueryXY
		KeyB *tKeyChainQueryXY
		tKeyChainQueryEmbeddedC
	}
)

func (dss decodeSuccessSuite) runKeyChainQueryStructSubtests(t *testing.T) {
	var (
		input = "keyA.keyX=val%20AX&keyA.keyY=val%20AY" +
			"&keyB.keyX=val%20BX&keyB.keyY=val%20BY" +
			"&keyC.keyX=val%20CX&keyC.keyY=val%20CY"

		runner = dss.withKeyChainSep('.')
	)

	runner.runTest(t, func(t *testing.T, decode tDecode) {
		var (
			originalAY, originalBY, originalCY = "orig AY", "orig BY", "orig CY"

			originalB = tKeyChainQueryXY{
				KeyX: "orig BX",
				KeyY: &originalBY,
			}
			target = tKeyChainQueryAB{
				KeyA: tKeyChainQueryXY{
					KeyX: "orig AX",
					KeyY: &originalAY,
				},
				KeyB: &originalB,
				tKeyChainQueryEmbeddedC: tKeyChainQueryEmbeddedC{
					KeyC: tKeyChainQueryXY{
						KeyX: "orig CX",
						KeyY: &originalCY,
					},
				},
			}
		)

		decode(input, &target)

		t.Run("KeyA.KeyX struct.string", func(t *testing.T) {
			assert.Equal(t, "val AX", target.KeyA.KeyX, "check target")
		})

		t.Run("KeyA.KeyY struct.*string", func(t *testing.T) {
			require.Equal(t, "val AY", *target.KeyA.KeyY, "check target")
			assert.Equal(t, "val AY", originalAY, "check original string")
		})

		t.Run("KeyB.KeyX *struct.string", func(t *testing.T) {
			require.Equal(t, "val BX", target.KeyB.KeyX, "check target")
			assert.Equal(t, "val BX", originalB.KeyX, "check original string")
		})

		t.Run("KeyB.KeyY *struct.*string", func(t *testing.T) {
			require.Equal(t, "val BY", *target.KeyB.KeyY, "check target")
			require.Equal(t, "val BY", *originalB.KeyY, "check original struct")
			assert.Equal(t, "val BY", originalBY, "check original string")
		})

		t.Run("KeyC.KeyX embedded struct.string", func(t *testing.T) {
			assert.Equal(t, "val CX", target.KeyC.KeyX, "check target")
		})

		t.Run("keyC.KeyY embedded struct.*string", func(t *testing.T) {
			require.Equal(t, "val CY", *target.KeyC.KeyY, "check target")
			assert.Equal(t, "val CY", originalCY, "check original string")
		})
	})
}

func (dss decodeSuccessSuite) runKeyChainQueryMixedSubtests(t *testing.T) {
	var (
		input  = "keyA.keyX=val%20AX"
		runner = dss.withKeyChainSep('.')
	)

	runner.runSubtest(t, "map[string]*struct{ KeyX *string }", func(t *testing.T, decode tDecode) {
		var (
			originalAX, originalBX = "orig AX", "orig BX"
			target                 = map[string]*struct{ KeyX *string }{
				"keyA": &struct{ KeyX *string }{KeyX: &originalAX},
				"keyB": &struct{ KeyX *string }{KeyX: &originalBX},
			}

			expectedAX, expectedBX = "val AX", "orig BX"
			expected               = map[string]*struct{ KeyX *string }{
				"keyA": &struct{ KeyX *string }{KeyX: &expectedAX},
				"keyB": &struct{ KeyX *string }{KeyX: &expectedBX},
			}
		)

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "val AX", originalAX)
		assert.Equal(t, "orig BX", originalBX)
	})

	runner.runSubtest(t, "struct{ KeyA *map[string]*string }", func(t *testing.T, decode tDecode) {
		var (
			originalAX, originalAY = "orig AX", "orig AY"
			originalA              = map[string]*string{"keyX": &originalAX, "keyY": &originalAY}
			target                 = struct{ KeyA *map[string]*string }{KeyA: &originalA}

			expectedAX, expectedAY = "val AX", "orig AY"
			expectedA              = map[string]*string{"keyX": &expectedAX, "keyY": &expectedAY}
			expected               = struct{ KeyA *map[string]*string }{KeyA: &expectedA}
		)

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "val AX", originalAX)
		assert.Equal(t, "orig AY", originalAY)
	})
}

// ----- Field
func (dss decodeSuccessSuite) runKeyChainFieldTests(t *testing.T) {
	// NOTE:
	// Just the basics for field level. Leave complex cases to the query level
	// for test brevity.

	var (
		input  = "keyA.keyX=val%20AX"
		runner = dss.withKeyChainSep('.')
	)

	runner.runSubtest(t, "basic map chain", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]map[string]string
			expected = map[string]map[string]string{
				"keyA": map[string]string{"keyX": "val AX"},
			}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})

	runner.runSubtest(t, "basic mixed chain", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]struct{ KeyX string }
			expected = map[string]struct{ KeyX string }{
				"keyA": struct{ KeyX string }{KeyX: "val AX"},
			}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})
}
