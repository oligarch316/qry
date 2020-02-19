package qry_test

import (
	"errors"
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

type testTrace struct{ *qry.TraceTree }

func (tt testTrace) log(t *testing.T) {
	if testing.Verbose() {
		t.Logf("Decode Trace:\n%s\n", tt.Sdump())
	}
}

type (
	decodeTest struct {
		configOpts      []qry.Option
		decodeLevel     qry.DecodeLevel
		input           string
		errorAssertions []func(*testing.T, error) bool
	}

	testOpt func(*decodeTest)
)

func newTest(opts ...testOpt) (res decodeTest) {
	for _, opt := range opts {
		opt(&res)
	}
	return
}

func (dt decodeTest) with(opts ...testOpt) decodeTest {
	res := decodeTest{
		input:           dt.input,
		decodeLevel:     dt.decodeLevel,
		errorAssertions: dt.errorAssertions,
	}

	if len(dt.configOpts) > 0 {
		res.configOpts = make([]qry.Option, len(dt.configOpts))
		copy(res.configOpts, dt.configOpts)
	}

	if len(dt.errorAssertions) > 0 {
		res.errorAssertions = make([]func(*testing.T, error) bool, len(dt.errorAssertions))
		copy(res.errorAssertions, dt.errorAssertions)
	}

	for _, opt := range opts {
		opt(&res)
	}

	return res
}

func (dt decodeTest) assert(t *testing.T, out interface{}) (testTrace, bool) {
	var (
		decoder = qry.NewDecoder(dt.configOpts...)
		trace   = testTrace{qry.NewTraceTree()}
	)

	actualErr := decoder.Decode(dt.decodeLevel, dt.input, out, trace)

	if len(dt.errorAssertions) < 1 {
		return trace, assert.NoError(t, actualErr)
	}

	success := true
	for _, check := range dt.errorAssertions {
		success = check(t, actualErr) && success
	}

	return trace, success
}

func (dt decodeTest) require(t *testing.T, out interface{}) testTrace {
	trace, success := dt.assert(t, out)
	if !success {
		trace.log(t)
		t.FailNow()
	}
	return trace
}

func configOptionsAs(opts ...qry.Option) testOpt {
	return func(d *decodeTest) { d.configOpts = opts }
}

func configOptionsAppend(opts ...qry.Option) testOpt {
	return func(d *decodeTest) { d.configOpts = append(d.configOpts, opts...) }
}

func inputAs(input string) testOpt {
	return func(d *decodeTest) { d.input = input }
}

func decodeLevelAs(level qry.DecodeLevel) testOpt {
	return func(d *decodeTest) { d.decodeLevel = level }
}

func checkDecodeError(assertions ...func(*testing.T, qry.DecodeError) bool) testOpt {
	check := func(t *testing.T, actual error) bool {
		var decodeErr qry.DecodeError
		if !errors.As(actual, &decodeErr) {
			return assert.Failf(t, "error is not a DecodeError", "%s (%T)", actual, actual)
		}

		success := true
		for _, assertion := range assertions {
			success = assertion(t, decodeErr) && success
		}
		return success
	}

	return func(d *decodeTest) { d.errorAssertions = append(d.errorAssertions, check) }
}

func assertDecodeLevel(level qry.DecodeLevel, msgAndArgs ...interface{}) func(*testing.T, qry.DecodeError) bool {
	return func(t *testing.T, actual qry.DecodeError) bool {
		return assert.Equal(t, level, actual.Level, msgAndArgs...)
	}
}

func checkStructFieldError(assertions ...func(*testing.T, qry.StructFieldError) bool) testOpt {
	check := func(t *testing.T, actual error) bool {
		var structFieldErr qry.StructFieldError
		if !errors.As(actual, &structFieldErr) {
			return assert.Failf(t, "error is not a StructFieldError", "%s (%T)", actual, actual)
		}

		success := true
		for _, assertion := range assertions {
			success = assertion(t, structFieldErr) && success
		}
		return success
	}

	return func(d *decodeTest) { d.errorAssertions = append(d.errorAssertions, check) }
}

func unwrapperOf(err error) (u interface{ Unwrap() error }, ok bool) {
	u, ok = err.(interface{ Unwrap() error })
	return
}

func errorMessageAs(errString string, msgAndArgs ...interface{}) testOpt {
	check := func(t *testing.T, actual error) bool {
		for unwrapper, ok := unwrapperOf(actual); ok; unwrapper, ok = unwrapperOf(actual) {
			actual = unwrapper.Unwrap()
		}
		return assert.EqualError(t, actual, errString, msgAndArgs...)
	}
	return func(d *decodeTest) { d.errorAssertions = append(d.errorAssertions, check) }
}

func errorMessageLike(errRx string, msgAndArgs ...interface{}) testOpt {
	check := func(t *testing.T, actual error) bool {
		for unwrapper, ok := unwrapperOf(actual); ok; unwrapper, ok = unwrapperOf(actual) {
			actual = unwrapper.Unwrap()
		}
		return assert.Regexp(t, errRx, actual.Error(), msgAndArgs...)
	}
	return func(d *decodeTest) { d.errorAssertions = append(d.errorAssertions, check) }
}
