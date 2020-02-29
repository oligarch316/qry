package qry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Error
func (des decodeErrorSuite) runKeyChainTests(t *testing.T) {
	t.Skip("TODO")

	// Unknown key
	// Non-indexable target type
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

	tKeyChainQueryAB struct {
		KeyA tKeyChainQueryXY
		KeyB *tKeyChainQueryXY
	}
)

func (dss decodeSuccessSuite) runKeyChainQueryStructSubtests(t *testing.T) {
	var (
		input  = "keyA.keyX=val%20AX&keyA.keyY=val%20AY&keyB.keyX=val%20BX&keyB.keyY=val%20BY"
		runner = dss.withKeyChainSep('.')
	)

	runner.runTest(t, func(t *testing.T, decode tDecode) {
		var (
			originalAY, originalBY = "orig AY", "orig BY"
			originalB              = tKeyChainQueryXY{
				KeyX: "orig BX",
				KeyY: &originalBY,
			}
			target = tKeyChainQueryAB{
				KeyA: tKeyChainQueryXY{
					KeyX: "orig AX",
					KeyY: &originalAY,
				},
				KeyB: &originalB,
			}
		)

		decode(input, &target)

		t.Run("KeyA.KeyX struct.string", func(t *testing.T) {
			assert.Equal(t, "val AX", target.KeyA.KeyX)
		})

		t.Run("KeyA.KeyY struct.*string", func(t *testing.T) {
			require.Equal(t, "val AY", *target.KeyA.KeyY)
			assert.Equal(t, "val AY", originalAY)
		})

		t.Run("KeyB.KeyX *struct.string", func(t *testing.T) {
			require.Equal(t, "val BX", target.KeyB.KeyX)
			assert.Equal(t, "val BX", originalB.KeyX)
		})

		t.Run("KeyB.KeyY *struct.*string", func(t *testing.T) {
			require.Equal(t, "val BY", *target.KeyB.KeyY)
			require.Equal(t, "val BY", *originalB.KeyY)
			assert.Equal(t, "val BY", originalBY)
		})
	})

	t.Run("various embedded struct|*struct tests", func(t *testing.T) {
		t.Skip("TODO")
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
	t.Skip("TODO")
}
