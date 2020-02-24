package qry_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const assertErrorAsFormat = `Error type not equal:
expected: %T
actual  : %T
message : %q`

type decodeErrorSuite struct{ decodeSuite }

func newErrorSuite(level qry.DecodeLevel) (res decodeErrorSuite) {
	res.level = level
	res.decodeOpts = []qry.Option{
		qry.SetLevelVia(level, qry.SetAllowLiteral),
	}
	return
}

func (des decodeErrorSuite) with(opts ...qry.Option) decodeErrorSuite {
	return decodeErrorSuite{des.decodeSuite.with(opts...)}
}

func (des decodeErrorSuite) assertMessage(t *testing.T, expected string, actual error, msgAndArgs ...interface{}) bool {
	for {
		if err := errors.Unwrap(actual); err != nil {
			actual = err
			continue
		}
		break
	}

	return assert.EqualError(t, actual, expected, msgAndArgs...)
}

func (des decodeErrorSuite) assertAs(t *testing.T, expected interface{}, actual error, msgAndArgs ...interface{}) bool {
	if !errors.As(actual, expected) {
		return assert.Fail(t, fmt.Sprintf(assertErrorAsFormat, expected, actual, actual), msgAndArgs...)
	}
	return true
}

func (des decodeErrorSuite) assertDecodeError(t *testing.T, expected string, actual error, msgAndArgs ...interface{}) bool {
	var decodeErr qry.DecodeError

	// Short circuit
	return assert.Error(t, actual, msgAndArgs...) &&
		des.assertAs(t, &decodeErr, actual, msgAndArgs...) &&
		des.assertMessage(t, expected, decodeErr, msgAndArgs...)
}

func (des decodeErrorSuite) assertStructFieldError(t *testing.T, expected string, actual error, msgAndArgs ...interface{}) bool {
	var (
		decodeErr      qry.DecodeError
		structFieldErr qry.StructFieldError
	)

	// Short circuit
	return assert.Error(t, actual, msgAndArgs...) &&
		des.assertAs(t, &decodeErr, actual, msgAndArgs...) &&
		des.assertAs(t, &structFieldErr, decodeErr, msgAndArgs...) &&
		des.assertMessage(t, expected, structFieldErr, msgAndArgs...)
}

// Helper for the very common "forced unescape error" subtest
func (des decodeErrorSuite) runUnescapeSubtest(t *testing.T, name string, test decodeSubtest) {
	var (
		errMsg = "forced unescape error"
		suite  = des.with(qry.ConvertUnescapeAs(func(_ string) (string, error) {
			return "", errors.New(errMsg)
		}))
	)

	suite.runSubtest(t, name, func(t *testing.T, decode tDecode) {
		test(t, func(input string, target interface{}) error {
			actual := decode(input, target)
			suite.assertDecodeError(t, errMsg, actual)
			return actual
		})
	})
}

// ----- Root errors
func (des decodeErrorSuite) runRootSubtests(t *testing.T) {
	des.runSubtest(t, "non-pointer target error", func(t *testing.T, decode tDecode) {
		var target string
		des.assertDecodeError(t, "non-pointer target", decode("xyz", target))
	})

	des.runSubtest(t, "nil pointer target error", func(t *testing.T, decode tDecode) {
		var target *string
		des.assertDecodeError(t, "nil pointer target", decode("xyz", target))
	})
}

// ----- Unsupported errors
func (des decodeErrorSuite) runUnsupportedCommonSubtests(t *testing.T) {
	var (
		input       = "xyz"
		expectedMsg = "TODO"
		noLitSuite  = des.with(qry.SetAllLevelsVia(qry.SetDisallowLiteral))
	)

	des.runSubtest(t, "chan target error", func(t *testing.T, decode tDecode) {
		var target chan struct{}
		des.assertDecodeError(t, expectedMsg, decode(input, &target))
	})

	des.runSubtest(t, "func target error", func(t *testing.T, decode tDecode) {
		var target func()
		des.assertDecodeError(t, expectedMsg, decode(input, &target))
	})

	noLitSuite.runSubtest(t, "literal string target error", func(t *testing.T, decode tDecode) {
		var target string
		des.assertDecodeError(t, expectedMsg, decode(input, &target))
	})
}

func (des decodeErrorSuite) runUnsupportedListContainerSubtests(t *testing.T) {
	des.runSubtest(t, "array target error", func(t *testing.T, decode tDecode) {
		t.Skip("TODO")
	})

	des.runSubtest(t, "slice target error", func(t *testing.T, decode tDecode) {
		t.Skip("TODO")
	})
}

func (des decodeErrorSuite) runUnsupportedKVContainerSubtests(t *testing.T) {
	des.runSubtest(t, "map target error", func(t *testing.T, decode tDecode) {
		t.Skip("TODO")
	})

	des.runSubtest(t, "struct target error", func(t *testing.T, decode tDecode) {
		t.Skip("TODO")
	})
}

// ----- Unmarshaler errors
type (
	tErrorUnmarshaler    struct{ forcedErr error }
	tErrorRawUnmarshaler struct{ forcedErr error }
)

func (teu tErrorUnmarshaler) UnmarshalText(text []byte) error        { return teu.forcedErr }
func (teru tErrorRawUnmarshaler) UnmarshalRawText(text []byte) error { return teru.forcedErr }

func (des decodeErrorSuite) runUnmarshalerSubtests(t *testing.T) {
	// Obligatory unescape test
	des.runUnescapeSubtest(t, "unescape error", func(t *testing.T, decodeAndAssert tDecode) {
		var target tUnmarshaler
		decodeAndAssert("xyz", &target)
	})

	var (
		forcedUnmarshalMsg = "forced unmarshal error"
		forcedUnmarshalErr = errors.New(forcedUnmarshalMsg)
	)

	des.runSubtest(t, "unmarshal text error", func(t *testing.T, decode tDecode) {
		target := tErrorUnmarshaler{forcedUnmarshalErr}
		des.assertDecodeError(t, forcedUnmarshalMsg, decode("xyz", &target))
	})

	des.runSubtest(t, "unmarshal raw text error", func(t *testing.T, decode tDecode) {
		target := tErrorRawUnmarshaler{forcedUnmarshalErr}
		des.assertDecodeError(t, forcedUnmarshalMsg, decode("xyz", &target))
	})
}

// ----- Literal errors
func (des decodeErrorSuite) runLiteralSubtests(t *testing.T) {
	// Obligatory unescape test
	des.runUnescapeSubtest(t, "unescape error", func(t *testing.T, decodeAndAssert tDecode) {
		var target string
		decodeAndAssert("xyz", &target)
	})

	var (
		input               = "xyz"
		expected, regexpErr = regexp.Compile(`strconv\..*: parsing .*: invalid syntax`)
	)

	require.NoError(t, regexpErr, "check expected regexp")

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
		{"bool", &boolTarget, true},

		{"int", &intTarget, true},
		{"int8", &int8Target, false},
		{"int16", &int16Target, false},
		{"int32", &int32Target, false},
		{"int64", &int64Target, false},

		{"uint", &uintTarget, true},
		{"uint8", &uint8Target, false},
		{"uint16", &uint16Target, false},
		{"uint32", &uint32Target, false},
		{"uint64", &uint64Target, false},

		{"float32", &float32Target, true},
		{"float64", &float64Target, false},

		{"complex64", &complex64Target, true},
		{"complex128", &complex128Target, false},
	}

	for _, item := range subtests {
		if !item.runOnShort && testing.Short() {
			continue
		}

		subtest := item
		des.runSubtest(t, subtest.name+" conversion", func(t *testing.T, decode tDecode) {
			// Perform decode
			actual := decode(input, subtest.target)

			// Require non-nil error
			require.Error(t, actual)

			var decodeErr qry.DecodeError

			// Require decode error
			if des.assertAs(t, &decodeErr, actual) {

				// Assert error message matches expectation format
				assert.Regexp(t, expected, decodeErr.Unwrap().Error())
			}
		})
	}
}

// ----- Faux literal errors
func (des decodeErrorSuite) runFauxLiteralSubtests(t *testing.T) {
	// Obligatory unescape test
	des.runUnescapeSubtest(t, "unescape error", func(t *testing.T, decodeAndAssert tDecode) {
		var target []rune
		decodeAndAssert("xyz", &target)
	})

	var (
		/*
		 * a       | 1 byte, 1 rune
		 * b       | 1 byte, 1 rune
		 * c       | 1 byte, 1 rune
		 * <space> | 1 byte, 1 rune
		 * 三 (sān) | 3 bytes, 1 rune
		 *
		 * Total   | 7 bytes, 5 runes
		 */

		input       = "abc%20三"
		expectedMsg = "insufficient destination array length"
	)

	des.runSubtest(t, "byte array too small error", func(t *testing.T, decode tDecode) {
		var target [6]byte
		des.assertDecodeError(t, expectedMsg, decode(input, &target))
	})

	des.runSubtest(t, "rune array too small error", func(t *testing.T, decode tDecode) {
		var target [4]rune
		des.assertDecodeError(t, expectedMsg, decode(input, &target))
	})
}
