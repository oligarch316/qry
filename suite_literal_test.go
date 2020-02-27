package qry_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===== Error
func (des decodeErrorSuite) runLiteralTests(t *testing.T) {
	des.runLiteralUnescapeSubtest(t)
	des.runLiteralConvertSubtests(t)
}

func (des decodeErrorSuite) runLiteralUnescapeSubtest(t *testing.T) {
	des.withUnescapeError("forced unescape error").runSubtest(t, "unescape error", func(t *testing.T, decode tDecode) {
		var target string
		actual := decode("xyz", &target)
		assertErrorMessage(t, "forced unescape error", actual)
	})
}

func (des decodeErrorSuite) runLiteralConvertSubtests(t *testing.T) {
	var (
		input    = "xyz"
		expected = regexp.MustCompile(`strconv\..*: parsing .*: invalid syntax`)
	)

	var (
		boolTarget bool

		intTarget   int
		int8Target  int8
		int16Target int16
		int32Target int32
		int64Target int64

		uintTarget   uint
		uint8Target  uint8
		uint16Target uint16
		uint32Target uint32
		uint64Target uint64

		float32Target float32
		float64Target float64

		complex64Target  complex64
		complex128Target complex128
	)

	subtests := []struct {
		name       string
		target     interface{}
		runOnShort bool
	}{
		{"bool", &boolTarget, true}, // Basic

		{"int", &intTarget, true},    // Basic
		{"int8", &int8Target, false}, // Extended ...
		{"int16", &int16Target, false},
		{"int32", &int32Target, false},
		{"int64", &int64Target, false},

		{"uint", &uintTarget, true},    // Basic
		{"uint8", &uint8Target, false}, // Extended ...
		{"uint16", &uint16Target, false},
		{"uint32", &uint32Target, false},
		{"uint64", &uint64Target, false},

		{"float32", &float32Target, true},  // Basic
		{"float64", &float64Target, false}, // Extended

		{"complex64", &complex64Target, true},    // Basic
		{"complex128", &complex128Target, false}, // Extended
	}

	for _, item := range subtests {
		if !item.runOnShort && testing.Short() {
			continue
		}

		subtest := item
		des.runSubtest(t, subtest.name+" conversion error", func(t *testing.T, decode tDecode) {
			actual := decode(input, subtest.target)
			assert.Regexp(t, expected, actual.Error())
		})
	}
}

// ===== Success
func (dss decodeSuccessSuite) runLiteralTests(t *testing.T) {
	dss.runLiteralStringSubtest(t)
	dss.runLiteralBoolSubtest(t)
	dss.runLiteralIntLikeSubtests(t)
	dss.runLiteralFloatLikeSubtests(t)
}

func (dss decodeSuccessSuite) runLiteralStringSubtest(t *testing.T) {
	dss.runSubtest(t, "string target", func(t *testing.T, decode tDecode) {
		var target string
		decode("abc%20xyz", &target)
		assert.Equal(t, "abc xyz", target)
	})
}

func (dss decodeSuccessSuite) runLiteralBoolSubtest(t *testing.T) {
	dss.runSubtest(t, "bool target", func(t *testing.T, decode tDecode) {
		var target bool
		decode("true", &target)
		assert.Equal(t, true, target)
	})
}

func (dss decodeSuccessSuite) runLiteralIntLikeSubtests(t *testing.T) {
	var (
		input    = "33"
		expected = 33
	)

	// Basic int
	dss.runSubtest(t, "int target", func(t *testing.T, decode tDecode) {
		var target int
		decode(input, &target)
		assert.Equal(t, expected, target)
	})

	// Extended int
	if !testing.Short() {
		dss.runSubtest(t, "int8 target", func(t *testing.T, decode tDecode) {
			var target int8
			decode(input, &target)
			assert.Equal(t, int8(expected), target)
		})

		dss.runSubtest(t, "int16 target", func(t *testing.T, decode tDecode) {
			var target int16
			decode(input, &target)
			assert.Equal(t, int16(expected), target)
		})

		dss.runSubtest(t, "int32 target", func(t *testing.T, decode tDecode) {
			var target int32
			decode(input, &target)
			assert.Equal(t, int32(expected), target)
		})

		dss.runSubtest(t, "int64 target", func(t *testing.T, decode tDecode) {
			var target int64
			decode(input, &target)
			assert.Equal(t, int64(expected), target)
		})
	}

	// Basic uint
	dss.runSubtest(t, "uint target", func(t *testing.T, decode tDecode) {
		var target uint
		decode(input, &target)
		assert.Equal(t, uint(expected), target)
	})

	// Extended uint
	if !testing.Short() {
		dss.runSubtest(t, "uint8 target", func(t *testing.T, decode tDecode) {
			var target uint8
			decode(input, &target)
			assert.Equal(t, uint8(expected), target)
		})

		dss.runSubtest(t, "uint16 target", func(t *testing.T, decode tDecode) {
			var target uint16
			decode(input, &target)
			assert.Equal(t, uint16(expected), target)
		})

		dss.runSubtest(t, "uint32 target", func(t *testing.T, decode tDecode) {
			var target uint32
			decode(input, &target)
			assert.Equal(t, uint32(expected), target)
		})

		dss.runSubtest(t, "uint64 target", func(t *testing.T, decode tDecode) {
			var target uint64
			decode(input, &target)
			assert.Equal(t, uint64(expected), target)
		})
	}
}

func (dss decodeSuccessSuite) runLiteralFloatLikeSubtests(t *testing.T) {
	var (
		input    = "2.718"
		expected = 2.718
	)

	// Basic float
	dss.runSubtest(t, "float32 target", func(t *testing.T, decode tDecode) {
		var target float32
		decode(input, &target)
		assert.Equal(t, float32(expected), target)
	})

	// Extended float
	if !testing.Short() {
		dss.runSubtest(t, "float64 target", func(t *testing.T, decode tDecode) {
			var target float64
			decode(input, &target)
			assert.Equal(t, float64(expected), target)
		})
	}

	// Basic complex
	dss.runSubtest(t, "complex64 target", func(t *testing.T, decode tDecode) {
		var target complex64
		decode(input, &target)
		assert.Equal(t, complex(float32(expected), 0), target)
	})

	// Extended complex
	if !testing.Short() {
		dss.runSubtest(t, "complex128 target", func(t *testing.T, decode tDecode) {
			var target complex128
			decode(input, &target)
			assert.Equal(t, complex(float64(expected), 0), target)
		})
	}
}
