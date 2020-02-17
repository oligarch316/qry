package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test commplex children like slices/maps/structs as direct interface elements
// are handled appropriately when the child level is in update mode.

type indirectSuite struct {
	replaceMode, updateMode decodeSuite

	defaultInput    string
	defaultExpected interface{}
	skipOnShort     bool
}

func newIndirectSuite(level qry.DecodeLevel, defaultInput string, defaultExpected interface{}, skipOnShort bool) indirectSuite {
	ds := decodeSuite{level: level}
	return indirectSuite{
		replaceMode: ds.with(qry.SetLevelVia(level, qry.SetReplaceIndirect, qry.SetAllowLiteral)),
		updateMode:  ds.with(qry.SetLevelVia(level, qry.SetUpdateIndirect, qry.SetAllowLiteral)),

		defaultInput:    defaultInput,
		defaultExpected: defaultExpected,
		skipOnShort:     skipOnShort,
	}
}

func (is indirectSuite) run(t *testing.T) {
	t.Run("replace", is.runReplaceSubtests)
	t.Run("update", is.runUpdateSubtests)
}

func (is indirectSuite) runReplaceSubtests(t *testing.T) {
	var (
		textInput    = "abc%20xyz"
		textExpected = "abc xyz"
	)

	if !testing.Short() || !is.skipOnShort {
		is.replaceMode.runSubtest(t, "non-zero *string target", func(t *testing.T, decode tDecode) {
			var (
				original = "orig"
				target   = &original
			)

			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, *target, "check target")
			assert.Equal(t, "orig", original, "check original")
		})

		is.replaceMode.runSubtest(t, "zero *string target", func(t *testing.T, decode tDecode) {
			var target *string
			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, *target)
		})

		is.replaceMode.runSubtest(t, "interface{non-zero *string} target", func(t *testing.T, decode tDecode) {
			var (
				original             = "orig"
				target   interface{} = &original
			)

			require.NoError(t, decode(is.defaultInput, &target))
			assert.Equal(t, is.defaultExpected, target, "check target")
			assert.Equal(t, "orig", original, "check original")
		})
	}

	is.replaceMode.runSubtest(t, "interface{} target", func(t *testing.T, decode tDecode) {
		var target interface{}
		require.NoError(t, decode(is.defaultInput, &target))
		assert.Equal(t, is.defaultExpected, target)
	})
}

func (is indirectSuite) runUpdateSubtests(t *testing.T) {
	var (
		textInput    = "abc%20xyz"
		textExpected = "abc xyz"
	)

	if !testing.Short() || !is.skipOnShort {
		is.updateMode.runSubtest(t, "interface{tNonPointerUnmarshaler} target", func(t *testing.T, decode tDecode) {
			var (
				imperativeVar             = "orig"
				modify                    = tNonPointerUnmarshaler(func(s string) { imperativeVar = s })
				target        interface{} = modify
			)

			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, imperativeVar)
		})

		is.updateMode.runSubtest(t, "non-zero *string target", func(t *testing.T, decode tDecode) {
			var (
				original = "orig"
				target   = &original
			)

			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, *target, "check target")
			assert.Equal(t, textExpected, original, "check original")
		})

		is.updateMode.runSubtest(t, "zero *string target", func(t *testing.T, decode tDecode) {
			var target *string
			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, *target)
		})

		is.updateMode.runSubtest(t, "interface{non-zero *string} target", func(t *testing.T, decode tDecode) {
			var (
				original             = "orig"
				target   interface{} = &original
			)

			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, *(target.(*string)), "check target")
			assert.Equal(t, textExpected, original, "check original")
		})

		is.updateMode.runSubtest(t, "interface{string} target", func(t *testing.T, decode tDecode) {
			var (
				tmp    string
				target interface{} = tmp
			)

			require.NoError(t, decode(textInput, &target))
			assert.Equal(t, textExpected, target)
		})
	}

	is.updateMode.runSubtest(t, "interface{} target", func(t *testing.T, decode tDecode) {
		var target interface{}
		require.NoError(t, decode(is.defaultInput, &target))
		assert.Equal(t, is.defaultExpected, target)
	})
}
