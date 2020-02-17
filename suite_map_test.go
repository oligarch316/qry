package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test commplex children like slices/maps/structs as direct map elements
// are handled appropriately when the child level is in update mode.

type mapSuite struct{ replaceMode, updateMode decodeSuite }

func newMapSuite(level qry.DecodeLevel) mapSuite {
	ds := decodeSuite{
		level:      level,
		decodeOpts: []qry.Option{qry.SetAllLevelsVia(qry.SetAllowLiteral)},
	}

	return mapSuite{
		replaceMode: ds.with(qry.SetLevelVia(level, qry.SetReplaceContainer)),
		updateMode:  ds.with(qry.SetLevelVia(level, qry.SetUpdateContainer)),
	}
}

func (ms mapSuite) run(t *testing.T) {
	t.Run("replace", ms.runReplaceSubtests)
	t.Run("update", ms.runUpdateSubtests)
}

func (ms mapSuite) runMulti(t *testing.T) { t.Run("multi", ms.runMultiSubtests) }

func (ms mapSuite) runReplaceSubtests(t *testing.T) {
	input := "key%20A=val%20A"

	ms.replaceMode.runSubtest(t, "non-zero map[string]non-zero *string target", func(t *testing.T, decode tDecode) {
		var (
			originalA = "orig A"
			originalB = "orig B"
			target    = map[string]*string{
				"key A": &originalA,
				"key B": &originalB,
			}

			expectedValA = "val A"
			expected     = map[string]*string{"key A": &expectedValA}
		)

		require.NoError(t, decode(input, &target))
		require.Equal(t, expected, target)
		assert.Equal(t, "orig A", originalA)
	})

	ms.replaceMode.runSubtest(t, "non-zero map[string]zero *string target", func(t *testing.T, decode tDecode) {
		var (
			target = map[string]*string{
				"key A": nil,
				"key B": nil,
			}

			expectedValA = "val A"
			expected     = map[string]*string{"key A": &expectedValA}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})

	ms.replaceMode.runSubtest(t, "zero map[string]string target", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]string
			expected = map[string]string{"key A": "val A"}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ms mapSuite) runUpdateSubtests(t *testing.T) {
	input := "key%20A=val%20A"

	ms.updateMode.runSubtest(t, "non-zero map[string]tNonPointerUnmarshaler target", func(t *testing.T, decode tDecode) {
		var (
			imperativeVarA = "orig A"
			imperativeVarB = "orig B"
			modifyA        = tNonPointerUnmarshaler(func(s string) { imperativeVarA = s })
			modifyB        = tNonPointerUnmarshaler(func(s string) { imperativeVarB = s })

			target = map[string]tNonPointerUnmarshaler{
				"key A": modifyA,
				"key B": modifyB,
			}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, "val A", imperativeVarA)
		assert.Equal(t, "orig B", imperativeVarB)
	})

	ms.updateMode.runSubtest(t, "non-zero map[string]non-zero *string target", func(t *testing.T, decode tDecode) {
		var (
			originalA = "orig A"
			originalB = "orig B"
			target    = map[string]*string{
				"key A": &originalA,
				"key B": &originalB,
			}

			expectedValA = "val A"
			expectedValB = "orig B"
			expected     = map[string]*string{
				"key A": &expectedValA,
				"key B": &expectedValB,
			}
		)

		require.NoError(t, decode(input, &target))
		require.Equal(t, expected, target)
		assert.Equal(t, "val A", originalA)
		assert.Equal(t, "orig B", originalB)
	})

	ms.updateMode.runSubtest(t, "non-zero map[string]zero *string target", func(t *testing.T, decode tDecode) {
		var (
			target = map[string]*string{
				"key A": nil,
				"key B": nil,
			}

			expectedValA = "val A"
			expected     = map[string]*string{
				"key A": &expectedValA,
				"key B": nil,
			}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})

	ms.updateMode.runSubtest(t, "zero map[string]string target", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]string
			expected = map[string]string{"key A": "val A"}
		)

		require.NoError(t, decode(input, &target))
		assert.Equal(t, expected, target)
	})
}

func (ms mapSuite) runMultiSubtests(t *testing.T) {
	// Just do one complex test to ensure multiple field logic is correct
	ms.updateMode.runSubtest(t, "update non-zero map[string]non-zero *string target", func(t *testing.T, decode tDecode) {
		var (
			input                                      = "key%20A=val%20A&key%20B=val%20B&key%20C=val%20C"
			originalA, originalB, originalC, originalD = "orig A", "orig B", "orig C", "orig D"
			target                                     = map[string]*string{
				"key A": &originalA,
				"key B": &originalB,
				"key C": &originalC,
				"key D": &originalD,
			}

			expectedA, expectedB, expectedC, expectedD = "val A", "val B", "val C", "orig D"
			expected                                   = map[string]*string{
				"key A": &expectedA,
				"key B": &expectedB,
				"key C": &expectedC,
				"key D": &expectedD,
			}
		)

		require.NoError(t, decode(input, &target))
		require.Equal(t, expected, target)
		assert.Equal(t, "val A", originalA)
		assert.Equal(t, "val B", originalB)
		assert.Equal(t, "val C", originalC)
		assert.Equal(t, "orig D", originalD)
	})
}
