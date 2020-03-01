package qry_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== General decode test runner
// This runner intends to store a decode level and decoder config options so
// as to supply individual tests with a curried decode() function.
//
// The runner manages creation of the test decoder and a trace for extra
// information on failure, and handles panics and trace printing via its
// provided curried decode() function.
//
// As a nod to the two "classes" of decode tests needed (success and error),
// the runner contains an error hook function for simple common error checking
// in the decode[Error|Success]Suite implementations.

type (
	tDecode    func(string, interface{}) error
	decodeTest func(*testing.T, tDecode)
)

type decodeRunner struct {
	errHook func(*testing.T, error) error
	level   qry.DecodeLevel
	opts    []qry.Option
}

func newDecodeRunner(level qry.DecodeLevel, errHook func(*testing.T, error) error) decodeRunner {
	return decodeRunner{
		errHook: errHook,
		level:   level,

		// Default config options for testing deviate from defaults for usage.
		// We'll relax "strict" errors in order to cover more code paths with
		// less config tweaking. The minority of tests designed to validate
		// said "strict" behavior are free to tweak.
		opts: []qry.Option{
			qry.SetAllLevelsVia(qry.SetAllowLiteral),
			qry.IgnoreInvalidKeys(true),
		},
	}
}

func (dr decodeRunner) with(opts ...qry.Option) decodeRunner {
	res := decodeRunner{errHook: dr.errHook, level: dr.level}
	if len(dr.opts) > 0 {
		res.opts = make([]qry.Option, len(dr.opts))
		copy(res.opts, dr.opts)
	}
	res.opts = append(res.opts, opts...)
	return res
}

func (dr decodeRunner) withSetOpts(setOpts ...qry.SetOption) decodeRunner {
	opts := make([]qry.Option, len(setOpts))
	for i, setOpt := range setOpts {
		opts[i] = qry.SetLevelVia(dr.level, setOpt)
	}
	return dr.with(opts...)
}

func (decodeRunner) dumpTraceOnFailure(t *testing.T, trace *qry.TraceTree) {
	if t.Failed() {
		t.Log(trace)
	}
}

func (dr decodeRunner) runTest(t *testing.T, test decodeTest) {
	// Setup
	var (
		decoder = qry.NewDecoder(dr.opts...)
		trace   = qry.NewTraceTree()
	)

	// Teardown
	if testing.Verbose() {
		// Use defer here to ensure trace dump in FailNow() situations
		defer dr.dumpTraceOnFailure(t, trace)
	}

	// Test
	assert.NotPanics(t, func() {
		test(t, func(input string, v interface{}) error {
			err := decoder.Decode(dr.level, input, v, trace)
			if dr.errHook == nil {
				return err
			}
			return dr.errHook(t, err)
		})
	})
}

func (dr decodeRunner) runSubtest(t *testing.T, name string, test decodeTest) {
	t.Run(name, func(t *testing.T) { dr.runTest(t, test) })
}

// ===== Error suite

type decodeErrorSuite struct{ decodeRunner }

func newDecodeErrorSuite(level qry.DecodeLevel) decodeErrorSuite {
	hook := func(t *testing.T, err error) error {
		require.Error(t, err)

		var decodeErr qry.DecodeError
		if !assertErrorAs(t, &decodeErr, err) {
			t.FailNow()
		}
		return decodeErr.Unwrap()
	}

	return decodeErrorSuite{newDecodeRunner(level, hook)}
}

// ----- Utility
const assertErrorAsFormat = `Error type not equal:
expected: %T
actual  : %T
message : %q`

func assertErrorMessage(t *testing.T, expected string, actual error, msgAndArgs ...interface{}) bool {
	for {
		if err := errors.Unwrap(actual); err != nil {
			actual = err
			continue
		}
		break
	}

	return assert.EqualError(t, actual, expected, msgAndArgs...)
}

func assertErrorAs(t *testing.T, expected interface{}, actual error, msgAndArgs ...interface{}) bool {
	if !errors.As(actual, expected) {
		return assert.Fail(t, fmt.Sprintf(assertErrorAsFormat, expected, actual, actual), msgAndArgs...)
	}
	return true
}

func (des decodeErrorSuite) withUnescapeError(msg string) decodeRunner {
	unescapeF := func(_ string) (string, error) { return "", errors.New(msg) }
	return des.with(qry.ConvertUnescapeAs(unescapeF))
}

// ----- Root errors
func (des decodeErrorSuite) runRootTests(t *testing.T) {
	input := "xyz"

	des.runSubtest(t, "non-pointer target error", func(t *testing.T, decode tDecode) {
		var target string
		actual := decode(input, target)
		assertErrorMessage(t, "non-pointer target", actual)
	})

	des.runSubtest(t, "nil pointer target error", func(t *testing.T, decode tDecode) {
		var target *string
		actual := decode(input, target)
		assertErrorMessage(t, "nil pointer target", actual)
	})
}

// ----- Unsupported errors
func (des decodeErrorSuite) runUnsupportedCommonTests(t *testing.T) {
	var (
		input    = "xyz"
		expected = "unsupported target type"
	)

	des.runSubtest(t, "chan target error", func(t *testing.T, decode tDecode) {
		var target chan struct{}
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "func target error", func(t *testing.T, decode tDecode) {
		var target func()
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})

	des.withSetOpts(qry.SetDisallowLiteral).runSubtest(t, "string target error", func(t *testing.T, decode tDecode) {
		var target string
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})
}

func (des decodeErrorSuite) runUnsupportedListTests(t *testing.T) {
	var (
		input    = "xyz"
		expected = "unsupported target type"
	)

	des.runSubtest(t, "array target error", func(t *testing.T, decode tDecode) {
		var target [5]string
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "slice target error", func(t *testing.T, decode tDecode) {
		var target []string
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})
}

func (des decodeErrorSuite) runUnsupportedKeyValTests(t *testing.T) {
	var (
		input    = "xyz"
		expected = "unsupported target type"
	)

	des.runSubtest(t, "map target error", func(t *testing.T, decode tDecode) {
		var target map[string]string
		actual := decode(input, &target)
		assertErrorMessage(t, expected, actual)
	})

	des.runSubtest(t, "struct target error", func(t *testing.T, decode tDecode) {
		var target struct{}
		actual := decode("xyz", &target)
		assertErrorMessage(t, expected, actual)
	})
}

// ===== Success suite

type decodeSuccessSuite struct{ decodeRunner }

func newDecodeSuccessSuite(level qry.DecodeLevel) (res decodeSuccessSuite) {
	hook := func(t *testing.T, err error) error {
		require.NoError(t, err)
		return nil
	}

	return decodeSuccessSuite{newDecodeRunner(level, hook)}
}

func (dss decodeSuccessSuite) withKeyChainSep(r rune) decodeRunner {
	return dss.with(qry.SeparateKeyChainBy(r))
}
