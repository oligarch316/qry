package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

// TODO: Test commplex children like slices/maps/structs as direct interface elements
// are handled appropriately when the child level is in update mode.

// ===== Success
func (dss decodeSuccessSuite) runIndirectCommonTests(t *testing.T) {
	t.Run("replace", dss.runIndirectCommonReplaceSubtests)
	t.Run("update", dss.runIndirectCommonUpdateSubtests)
}

func (dss decodeSuccessSuite) runIndirectDefaultTests(t *testing.T, input string, expected interface{}) {
	t.Run("default", func(t *testing.T) {
		dss.runSubtest(t, "interface{} target", func(t *testing.T, decode tDecode) {
			var target interface{}
			decode(input, &target)
			assert.Equal(t, expected, target)
		})

		dss.withSetOpts(qry.SetReplaceIndirect).runSubtest(t, "replace interface{non-zero *string} target", func(t *testing.T, decode tDecode) {
			var (
				original             = "orig"
				target   interface{} = &original
			)

			decode(input, &target)
			assert.Equal(t, expected, target, "check target")
			assert.Equal(t, "orig", original, "check original")
		})
	})
}

func (dss decodeSuccessSuite) runIndirectCommonReplaceSubtests(t *testing.T) {
	var (
		textInput    = "abc%20xyz"
		textExpected = "abc xyz"
		runner       = dss.withSetOpts(qry.SetReplaceIndirect)
	)

	runner.runSubtest(t, "non-zero *string target", func(t *testing.T, decode tDecode) {
		var (
			original = "orig"
			target   = &original
		)

		decode(textInput, &target)
		assert.Equal(t, textExpected, *target, "check target")
		assert.Equal(t, "orig", original, "check original")
	})

	runner.runSubtest(t, "zero *string target", func(t *testing.T, decode tDecode) {
		var target *string
		decode(textInput, &target)
		assert.Equal(t, textExpected, *target)
	})
}

func (dss decodeSuccessSuite) runIndirectCommonUpdateSubtests(t *testing.T) {
	var (
		textInput    = "abc%20xyz"
		textExpected = "abc xyz"
		runner       = dss.withSetOpts(qry.SetUpdateIndirect)
	)

	runner.runSubtest(t, "interface{tNonPointerUnmarshaler} target", func(t *testing.T, decode tDecode) {
		var (
			imperativeVar             = "orig"
			modify                    = tNonPointerUnmarshaler(func(s string) { imperativeVar = s })
			target        interface{} = modify
		)

		decode(textInput, &target)
		assert.Equal(t, textExpected, imperativeVar)
	})

	runner.runSubtest(t, "non-zero *string target", func(t *testing.T, decode tDecode) {
		var (
			original = "orig"
			target   = &original
		)

		decode(textInput, &target)
		assert.Equal(t, textExpected, *target, "check target")
		assert.Equal(t, textExpected, original, "check original")
	})

	runner.runSubtest(t, "zero *string target", func(t *testing.T, decode tDecode) {
		var target *string
		decode(textInput, &target)
		assert.Equal(t, textExpected, *target)
	})

	runner.runSubtest(t, "interface{non-zero *string} target", func(t *testing.T, decode tDecode) {
		var (
			original             = "orig"
			target   interface{} = &original
		)

		decode(textInput, &target)
		assert.Equal(t, textExpected, *(target.(*string)), "check target")
		assert.Equal(t, textExpected, original, "check original")
	})

	runner.runSubtest(t, "interface{string} target", func(t *testing.T, decode tDecode) {
		var (
			tmp    string
			target interface{} = tmp
		)

		decode(textInput, &target)
		assert.Equal(t, textExpected, target)
	})
}
