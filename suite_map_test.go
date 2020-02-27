package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test commplex children like slices/maps/structs as direct map elements
// are handled appropriately when the child level is in update mode.

// Success
func (dss decodeSuccessSuite) runMapSingleTests(t *testing.T) {
	t.Run("replace", dss.runMapSingleReplaceSubtests)
	t.Run("update", dss.runMapSingleUpdateSubtests)
}

func (dss decodeSuccessSuite) runMapMultiTests(t *testing.T) {
	// Just do one complex update test to ensure multiple field logic is correct
	subtest := func(t *testing.T, decode tDecode) {
		var (
			input = "key%20A=val%20A&key%20B=val%20B&key%20C=val%20C"

			originalA, originalB, originalC, originalD = "orig A", "orig B", "orig C", "orig D"
			expectedA, expectedB, expectedC, expectedD = "val A", "val B", "val C", "orig D"

			target = map[string]*string{
				"key A": &originalA,
				"key B": &originalB,
				"key C": &originalC,
				"key D": &originalD,
			}

			expected = map[string]*string{
				"key A": &expectedA,
				"key B": &expectedB,
				"key C": &expectedC,
				"key D": &expectedD,
			}
		)

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "val A", originalA)
		assert.Equal(t, "val B", originalB)
		assert.Equal(t, "val C", originalC)
		assert.Equal(t, "orig D", originalD)
	}

	dss.withSetOpts(qry.SetUpdateContainer).runSubtest(t, "update non-zero map[string]non-zero *string target", subtest)
}

func (dss decodeSuccessSuite) runMapSingleReplaceSubtests(t *testing.T) {
	var (
		input  = "key%20A=val%20A"
		runner = dss.withSetOpts(qry.SetReplaceContainer)
	)

	runner.runSubtest(t, "non-zero map[string]non-zero *string target", func(t *testing.T, decode tDecode) {
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

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "orig A", originalA)
	})

	runner.runSubtest(t, "non-zero map[string]zero *string target", func(t *testing.T, decode tDecode) {
		var (
			target = map[string]*string{
				"key A": nil,
				"key B": nil,
			}

			expectedValA = "val A"
			expected     = map[string]*string{"key A": &expectedValA}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})

	runner.runSubtest(t, "zero map[string]string target", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]string
			expected = map[string]string{"key A": "val A"}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})
}

func (dss decodeSuccessSuite) runMapSingleUpdateSubtests(t *testing.T) {
	var (
		input  = "key%20A=val%20A"
		runner = dss.withSetOpts(qry.SetUpdateContainer)
	)

	runner.runSubtest(t, "non-zero map[string]tNonPointerUnmarshaler target", func(t *testing.T, decode tDecode) {
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

		decode(input, &target)
		assert.Equal(t, "val A", imperativeVarA)
		assert.Equal(t, "orig B", imperativeVarB)
	})

	runner.runSubtest(t, "non-zero map[string]non-zero *string target", func(t *testing.T, decode tDecode) {
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

		decode(input, &target)
		require.Equal(t, expected, target)
		assert.Equal(t, "val A", originalA)
		assert.Equal(t, "orig B", originalB)
	})

	runner.runSubtest(t, "non-zero map[string]zero *string target", func(t *testing.T, decode tDecode) {
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

		decode(input, &target)
		assert.Equal(t, expected, target)
	})

	runner.runSubtest(t, "zero map[string]string target", func(t *testing.T, decode tDecode) {
		var (
			target   map[string]string
			expected = map[string]string{"key A": "val A"}
		)

		decode(input, &target)
		assert.Equal(t, expected, target)
	})
}
